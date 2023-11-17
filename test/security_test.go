package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/messaging"
	"gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
)

// This is a transaction hash that exists on preprpod that is for a different CP
const OldValidTransactionHash = "ce964014ea9c4b6ab884f82592846fde0c652a652db63f06dd549e78d9d78f86"

type TestHarness struct {
	suite.Suite
	sync.WaitGroup
	cleanup func(context.Context) error
}

func TestSecurity(t *testing.T) {
	s := new(TestHarness)
	suite.Run(t, s)
}

func (s *TestHarness) SetupSuite() {
	SetupDMSTestingConfiguration("target", 0);
	OnboardTestComputeProvider()
	RunTestComputeProvider()

	s.cleanup = tracing.InitTracer()
	s.WaitGroup.Add(1)
}

func (s *TestHarness) TearDownSuite() {
	s.WaitGroup.Done()
	s.cleanup(context.Background())
	os.RemoveAll("/tmp/nunet-target")
	os.RemoveAll("/tmp/nunet-client")
}

func DefaultDeploymentRequest(tx_hash string) models.DeploymentRequest {
	var req models.DeploymentRequest
	// Hash of a valid job that has a valid datum and payment, but that doesn't list this CP DMS's address as the chosen CP
	req.TxHash = tx_hash
	req.RequesterWalletAddress = "addr_test1qrrysjx7gg6e2h8qvsqc29lg37yttq6cnww72637sdgfm7c58xcwszypj5fz8mmdvkv2a7wew2tthvvftj02gdeaf4vsc849la"
	req.MaxNtx = 2
	req.Blockchain = "Cardano"
	req.ServiceType = "ml-training-cpu"
	req.Timestamp = time.Now()
	req.Params.MachineType = "cpu"

	// Simple Model URL
	req.Params.ModelURL = "https://gist.github.com/luigy/d63eec5cb33d9f789969fafe04ee3ae9"

	// Valid cpu image ID
	req.Params.ImageID = "registry.gitlab.com/nunet/ml-on-gpu/ml-on-cpu-service/develop/ml-on-cpu"

	req.Params.RemoteNodeID = "invalidremoteid"
	req.Params.LocalNodeID = "invalidlocalid"
	req.Params.LocalPublicKey = "invalidpublickey"

	// Low complexity and constraints
	req.Constraints.Complexity = "Low"
	req.Constraints.CPU = 1500
	req.Constraints.RAM = 2000
	req.Constraints.Vram = 2000
	req.Constraints.Power = 170
	req.Constraints.Time = 1

	return req;
}

// Test if the the CP DMS will run a undervalued job
func (s *TestHarness) TestTxUndervaluation() {
	spClient, err := CreateServiceProviderTestingClient(s)
	s.Nil(err, "Failed to create testing client");

	// Create a request with very high constraints but with the minimum NTX
	req := DefaultDeploymentRequest(OldValidTransactionHash)
	req.MaxNtx = 1;
	req.Constraints.CPU = 1000000;
	req.Constraints.RAM = 2000000;
	req.Constraints.Vram = 2000000;
	req.Constraints.Power = 20000000;
	req.Constraints.Time = 10000000;

	spClient.SendDeploymentRequest(req)

	spClient.AssertJobFail("The CP DMS ran the job even though the requirements are too high for the payment")
}

// Convenience function to check for job failure and if the job runs respond with a custom error message.
func ( spClient *SPTestClient ) AssertJobFail( job_ran_error string ) {
	s := spClient.s

	firstUpdate, err := spClient.GetNextDeploymentUpdate()
	s.Nil(err, "Failed to get first deployment update")
	s.NotEqual(libp2p.MsgJobStatus, firstUpdate.MsgType, job_ran_error)

	// If we have received a message a job is running, lets also confirm that the DeploymentRequest is successful as well
	if firstUpdate.MsgType == libp2p.MsgJobStatus {
		secondUpdate, err := spClient.GetNextDeploymentUpdate()
		s.Equal(libp2p.MsgDepResp, secondUpdate.MsgType, "We didn't receive a DeploymentResponse for our DeploymentRequest")
		s.Nil(err, "Failed to get second deployment update")
		s.Equal(false, secondUpdate.Response.Success, job_ran_error)
	}
}

// Test that the CP DMS will not run a job with an invalid tx hash
func (s *TestHarness) TestTxHashValidation() {
	spClient, err := CreateServiceProviderTestingClient(s)
	s.Nil(err, "Failed to create testing client");

	req := DefaultDeploymentRequest("invalidtxhash")

	spClient.SendDeploymentRequest(req)

	spClient.AssertJobFail("Malicious SP sent an invalid tx-hash but the CP confirmed deployment success")
}

// Test that the CP DMS will not run a job with a valid tx hash that does not have the correct amount of NTX
func (s *TestHarness) TestTxNTXValidation() {
	spClient, err := CreateServiceProviderTestingClient(s)
	s.Nil(err, "Failed to create testing client");

	req := DefaultDeploymentRequest(OldValidTransactionHash)
	req.MaxNtx = 1000000

	spClient.SendDeploymentRequest(req)

	spClient.AssertJobFail("Malicious SP sent a valid transaction but decalred a higher payout to the DMS and the job ran success")
}

func (s *TestHarness) TestModelValidation() {
	spClient, err := CreateServiceProviderTestingClient(s)
	s.Nil(err, "Failed to create testing client");

	req := DefaultDeploymentRequest(OldValidTransactionHash)
	req.Params.ModelURL = "null"

	spClient.SendDeploymentRequest(req)

	spClient.AssertJobFail("Malicious SP sent an invalid model url and CP still tried to run it")
}

func GenerateTestKeyPair() (crypto.PrivKey, error) {
	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519, // Ed25519 are nice short
		-1,             // No length required for Ed25519
	)

	return priv, err
}

func SetupDMSTestingConfiguration( tempDirectoryName string, port int ) {
	dmsTempDir := fmt.Sprintf("/tmp/nunet-%s", tempDirectoryName)
	os.Mkdir(dmsTempDir, 0755)
	ioutil.WriteFile(fmt.Sprintf("%s/metadataV2.json", dmsTempDir), []byte(testMetadata), 0644)
	config.LoadConfig()
	addrs := [2]string{fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port), fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port)}
	config.SetConfig("general.metadata_path", dmsTempDir)
	config.SetConfig("p2p.listen_address", addrs)
}

func OnboardTestComputeProvider() {
	db.ConnectDatabase(afero.NewOsFs())

	availableResources := models.AvailableResources{
		TotCpuHz:  int(2000000),
		CpuNo:     int(32),
		CpuHz:     2000,
		PriceCpu:  0,
		Ram:       int(21334),
		PriceRam:  0,
		Vcpu:      int(math.Floor((float64(2000000)) / 2000)),
		Disk:      0,
		PriceDisk: 0,
	}

	freeResources := models.FreeResources{
		ID: 1,
		TotCpuHz: availableResources.TotCpuHz,
		PriceCpu: 0,
		Ram: availableResources.Ram,
		PriceRam: 0,
		Vcpu: 0,
		Disk: 0,
		PriceDisk: 0,
	}

	db.DB.Create(&availableResources)
	db.DB.Create(&freeResources)
}

func RunTestComputeProvider() error {
	// Run deployment worker so that deployment requests are handled
	go messaging.DeploymentWorker()

	// Create a new key pair for Node Id Generation
	pair, err := GenerateTestKeyPair()

	runAsServer := true;
	libp2p.RunNode(pair, runAsServer)
	return err
}

// Interface for SP to communicate with testing CP
type SPTestClient struct
{
	reader *bufio.Reader
	writer *bufio.Writer
	s *TestHarness
}

// Convenience type for DeploymentUpdate with unmarshalled data
type CPUpdate struct
{
	MsgType string
	Response models.DeploymentResponse
	Services models.Services
}

func CreateServiceProviderTestingClient( s *TestHarness ) (SPTestClient, error) {
	ctx := context.Background()

	SetupDMSTestingConfiguration("client", 0)

	pair, err := GenerateTestKeyPair()
	host, dht, err := libp2p.NewHost(ctx, pair, false)
	p2p := *libp2p.DMSp2pInit(host, dht)
	err = p2p.BootstrapNode(ctx)

	cpID := GetTestComputeProviderID()
	stream, err := host.NewStream(context.Background(), cpID, libp2p.DepReqProtocolID)

	reader := bufio.NewReader(stream)
	writer := bufio.NewWriter(stream)

	return SPTestClient{reader,writer,s}, err
}

func GetTestComputeProviderID() peer.ID {
	p2p := libp2p.GetP2P()
	return p2p.Host.ID()
}

func ( client *SPTestClient ) SendDeploymentRequest( request models.DeploymentRequest ) error {
	msg, json_err := json.Marshal(request)
	if json_err != nil {
		return json_err;
	}

	_, write_err := client.writer.WriteString(fmt.Sprintf("%s\n", msg))
	if write_err != nil {
		return write_err;
	}

	flush_err := client.writer.Flush()
	if flush_err != nil {
		return flush_err;
	}

	return nil
}

func ( client *SPTestClient ) GetNextDeploymentUpdate() (CPUpdate, error) {
	var update CPUpdate;

	msg, err := client.reader.ReadString('\n')
	if err != nil {
		return update, err
	}

	depUpdate := models.DeploymentUpdate{}
	err = json.Unmarshal([]byte(msg), &depUpdate)
	if err != nil {
		return update, err
	}

	update = CPUpdate{
		MsgType: depUpdate.MsgType,
	}

	switch depUpdate.MsgType {
	case libp2p.MsgDepResp:
		err = json.Unmarshal([]byte(depUpdate.Msg), &update.Response)
		if err != nil {
			return update, err
		}
	case libp2p.MsgJobStatus:
		json.Unmarshal([]byte(depUpdate.Msg), &update.Services)
		if err != nil {
			return update, err
		}
	default:
		return update, errors.New(fmt.Sprintf("Invalid Message Type %s", depUpdate.MsgType))
	}

	return update, nil
}

// Generated metadata for testing
var testMetadata string = `
{
 "update_timestamp": 1698332902,
 "resource": {
  "memory_max": 31674,
  "total_core": 16,
  "cpu_max": 67198
 },
 "available": {
  "cpu": 42942,
  "memory": 10340
 },
 "reserved": {
  "cpu": 24256,
  "memory": 21334
 },
 "network": "nunet-team",
 "public_key": "addr_test1qrrysjx7gg6e2h8qvsqc29lg37yttq6cnww72637sdgfm7c58xcwszypj5fz8mmdvkv2a7wew2tthvvftj02gdeaf4vsc849la",
 "allow_cardano": true
}`
