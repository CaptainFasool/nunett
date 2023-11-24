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
	"gitlab.com/nunet/device-management-service/integrations/oracle"
	"gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
)

// This is a transaction hash that exists on preprpod
const OldValidTransactionHash = "723fb1e2260c8f7c31b5760177c365917fcb39291925ce8b5c897c51d0de8fe7"

// Address of the SP, currently same as CP
const RequesterAddress = "addr_test1vzgxkngaw5dayp8xqzpmajrkm7f7fleyzqrjj8l8fp5e8jcc2p2dk"

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

func ( spClient *SPTestClient ) DefaultDeploymentRequest(tx_hash string) models.DeploymentRequest {
	var req models.DeploymentRequest
	// Hash of a valid job that has a valid datum and payment, but that doesn't list this CP DMS's address as the chosen CP
	req.TxHash = tx_hash
	req.RequesterWalletAddress = RequesterAddress
	req.MaxNtx = 2
	req.Blockchain = "Cardano"
	req.ServiceType = "ml-training-cpu"
	req.Timestamp = time.Now()
	req.Params.MachineType = "cpu"

	// Simple Model URL
  req.Params.ModelURL = basicTensorflowModelURL

	// Valid cpu image ID
	req.Params.ImageID = "registry.gitlab.com/nunet/ml-on-gpu/ml-on-cpu-service/develop/ml-on-cpu"

	cpId := GetTestComputeProviderID()

	computeProviderPubKey, err := libp2p.GetP2P().Host.Peerstore().PubKey(cpId).Raw()
	spClient.s.Nil(err, "Failed to obtain compute provider public key");

	// NOTE (divam): currently the public key is added to JSON without the typical base64 encoding
	req.Params.RemoteNodeID = cpId.String()
	req.Params.RemotePublicKey = string(computeProviderPubKey)
	req.Params.LocalNodeID = spClient.selfID.String()
	selfPubKey, _ := spClient.selfPublicKey.Raw()
	req.Params.LocalPublicKey = string(selfPubKey)

	// Low complexity and constraints
	req.Constraints.Complexity = "Low"
	req.Constraints.CPU = 1500
	req.Constraints.RAM = 2000
	req.Constraints.Vram = 2000
	req.Constraints.Power = 170
	req.Constraints.Time = 1

	oracleResp, err := oracle.FundContractRequest(&oracle.FundingRequest{
		ServiceProviderAddr: req.RequesterWalletAddress,
		// TODO: obtain from CP metadata
		ComputeProviderAddr: "addr_test1vzgxkngaw5dayp8xqzpmajrkm7f7fleyzqrjj8l8fp5e8jcc2p2dk",
		EstimatedPrice:      int64(req.MaxNtx),
	})
	spClient.s.Nil(err, "Failed to obtain oracleResp");

	req.MetadataHash = oracleResp.MetadataHash
	req.WithdrawHash = oracleResp.WithdrawHash
	req.RefundHash = oracleResp.RefundHash
	req.Distribute_50Hash = oracleResp.Distribute_50Hash
	req.Distribute_75Hash = oracleResp.Distribute_75Hash

	return req;
}

// Test if the the CP DMS will run a undervalued job
func (s *TestHarness) TestTxUndervaluation() {
	spClient, err := CreateServiceProviderTestingClient(s)
	s.Nil(err, "Failed to create testing client");

	// Create a request with very high constraints but with the minimum NTX
	req := spClient.DefaultDeploymentRequest(OldValidTransactionHash)
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

		if secondUpdate.Response.Success {
			// Wait for CP to finish the current job, as only one job can run at a time
			// This is to prevent test failure due to "depReq already in progress. Refusing to accept."
			for {
				update, err := spClient.GetNextDeploymentUpdate()
				s.Nil(err, "Failed to get deployment update")
				if update.MsgType == libp2p.MsgLogStdout || update.MsgType == libp2p.MsgLogStderr {
					continue
				}
				// Final confirmation that the job finished
				s.Equal(libp2p.MsgJobStatus, update.MsgType, "Expected Job Status update")
				break;
			}
		}
	}
}

// Test that the CP DMS will not run a job with an invalid tx hash
func (s *TestHarness) TestTxHashValidation() {
	spClient, err := CreateServiceProviderTestingClient(s)
	s.Nil(err, "Failed to create testing client");

	req := spClient.DefaultDeploymentRequest("invalidtxhash")

	spClient.SendDeploymentRequest(req)

	spClient.AssertJobFail("Malicious SP sent an invalid tx-hash but the CP confirmed deployment success")
}

// Test that the CP DMS will not run a job with a valid tx hash that does not have the correct amount of NTX
func (s *TestHarness) TestTxNTXValidation() {
	spClient, err := CreateServiceProviderTestingClient(s)
	s.Nil(err, "Failed to create testing client");

	req := spClient.DefaultDeploymentRequest(OldValidTransactionHash)
	req.MaxNtx = 1000000

	spClient.SendDeploymentRequest(req)

	spClient.AssertJobFail("Malicious SP sent a valid transaction but decalred a higher payout to the DMS and the job ran success")
}

// Test that the CP DMS will only run the job when the Params specifying a correct RemotePublicKey
func (s *TestHarness) TestValidCPPublicKey() {
	spClient, err := CreateServiceProviderTestingClient(s)
	s.Nil(err, "Failed to create testing client");

	req := spClient.DefaultDeploymentRequest(OldValidTransactionHash)
	req.Params.RemotePublicKey = "invalid-remote-key"

	spClient.SendDeploymentRequest(req)

	spClient.AssertJobFail("Malicious SP sent a DeploymentRequest with invalid RemotePublicKey")
}

// Test that the CP DMS will only run the job when the Params specifying a correct RemoteNodeID
func (s *TestHarness) TestValidCPNodeId() {
	spClient, err := CreateServiceProviderTestingClient(s)
	s.Nil(err, "Failed to create testing client");

	req := spClient.DefaultDeploymentRequest(OldValidTransactionHash)
	req.Params.RemoteNodeID = "invalid-remote-id"

	spClient.SendDeploymentRequest(req)

	spClient.AssertJobFail("Malicious SP sent a DeploymentRequest with invalid RemoteNodeID")
}

type OracleRewardReqPayload struct
{
	JobStatus            string // whether job is running or exited; one of these 'running', 'finished without errors', 'finished with errors'
	JobDuration          int64  // job duration in minutes
	EstimatedJobDuration int64  // job duration in minutes
	LogURL               string
}

func ( spClient *SPTestClient ) getSignaturesFromOracle(req models.DeploymentRequest, payload OracleRewardReqPayload) (oracleResp *oracle.RewardResponse) {
	oracleResp, err := oracle.Oracle.WithdrawTokenRequest(&oracle.RewardRequest{
		JobStatus:            payload.JobStatus,
		JobDuration:          payload.JobDuration,
		EstimatedJobDuration: payload.EstimatedJobDuration,
		LogPath:              payload.LogURL,
		MetadataHash:         req.MetadataHash,
		WithdrawHash:         req.WithdrawHash,
		RefundHash:           req.RefundHash,
		Distribute_50Hash:    req.Distribute_50Hash,
		Distribute_75Hash:    req.Distribute_75Hash,
	})

	spClient.s.Nil(err, "Failed to obtain signatures from oracle");

	return oracleResp
}

func GenerateTestKeyPair() (crypto.PrivKey, crypto.PubKey, error) {
	priv, pub, err := crypto.GenerateKeyPair(
		crypto.Ed25519, // Ed25519 are nice short
		-1,             // No length required for Ed25519
	)

	return priv, pub, err
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
	pair, _, err := GenerateTestKeyPair()

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
	selfID peer.ID
	selfPublicKey crypto.PubKey
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

	pair, selfPublicKey, err := GenerateTestKeyPair()
	host, dht, err := libp2p.NewHost(ctx, pair, false)
	p2p := *libp2p.DMSp2pInit(host, dht)
	selfID := p2p.Host.ID()
	err = p2p.BootstrapNode(ctx)

	cpID := GetTestComputeProviderID()
	stream, err := host.NewStream(context.Background(), cpID, libp2p.DepReqProtocolID)

	reader := bufio.NewReader(stream)
	writer := bufio.NewWriter(stream)
	return SPTestClient{reader,writer,s, selfID, selfPublicKey}, err
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
	case libp2p.MsgLogStdout:
	case libp2p.MsgLogStderr:
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
 "public_key": "addr_test1vzgxkngaw5dayp8xqzpmajrkm7f7fleyzqrjj8l8fp5e8jcc2p2dk",
 "allow_cardano": true
}`

// Prints 'GPU found' / 'No GPU found'
var basicTensorflowModelURL = "https://gist.githubusercontent.com/luigy/d63eec5cb33d9f789969fafe04ee3ae9/raw/c9722361c24e7520e5ebc084f94358fc0858753e/tesorflow.py"

// Prints 'GPU found' / 'No GPU found' and sleep for 1 min
var oneMinSleepModelURL = "https://gist.githubusercontent.com/dfordivam/10fa4b73f1d51cfc0d94aea844634bf7/raw/01490c73a7e988a11484b0ed42946cb3422b570a/one-min-sleep.py"

// Prints 'GPU found' / throws error on CPU
var gpuOnlyModelURL = "https://gist.githubusercontent.com/dfordivam/703232aafb3ad3348095a0890cfe7911/raw/589014937e38d6d80f31c15d24b18b33a287d735/gistfile1.txt"
