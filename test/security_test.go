package main

import (
	"io/ioutil"
	"bufio"
	"math"
	"fmt"
	"context"
	"encoding/json"
	"testing"
	"time"
	"sync"
	"os"
	"errors"

	"github.com/stretchr/testify/suite"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/internal/messaging"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/crypto"
)

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
	SetupDMSTestingConfiguration("target", 9123);
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

func (s *TestHarness) TestTxHashValidation() {
	spClient, err := CreateServiceProviderTestingClient()
	s.Nil(err, "Failed to create testing client");

	var req models.DeploymentRequest
	req.TxHash = "notavalidhash"
	req.RequesterWalletAddress = "addr_test1qrrysjx7gg6e2h8qvsqc29lg37yttq6cnww72637sdgfm7c58xcwszypj5fz8mmdvkv2a7wew2tthvvftj02gdeaf4vsc849la"
	req.MaxNtx = 2
	req.Blockchain = "Cardano"
	req.ServiceType = "ml-training-cpu"
	req.Timestamp = time.Now()
	req.Params.MachineType = "cpu"
	req.Params.ModelURL = "https://gist.github.com/luigy/d63eec5cb33d9f789969fafe04ee3ae9"
	req.Params.ImageID = "registry.gitlab.com/nunet/ml-on-gpu/ml-on-cpu-service/develop/ml-on-cpu"
	req.Params.RemoteNodeID = "invalidremoteid"
	req.Params.LocalNodeID = "invalidlocalid"
	req.Params.LocalPublicKey = "invalidpublickey"
	req.Constraints.Complexity = "Low"
	req.Constraints.CPU = 1500
	req.Constraints.RAM = 2000
	req.Constraints.Vram = 2000
	req.Constraints.Power = 170
	req.Constraints.Time = 1

	spClient.SendDeploymentRequest(req)

	firstUpdate, err := spClient.GetNextDeploymentUpdate()
	s.Nil(err, "Failed to get first deployment update")
	s.NotEqual(libp2p.MsgJobStatus, firstUpdate.MsgType, "Malicious SP send an invalid tx-hash but the CP started the job")

	// If we have received a message a job is running, lets also confirm that the DeploymentRequest is successful as well
	if firstUpdate.MsgType == libp2p.MsgJobStatus {
		secondUpdate, err := spClient.GetNextDeploymentUpdate()
		s.Equal(libp2p.MsgDepResp, secondUpdate.MsgType, "We didn't receive a DeploymentResponse for our DeploymentRequest")
		s.Nil(err, "Failed to get second deployment update")
		s.Equal(false, secondUpdate.Response.Success, "Malicious SP sent an invalid tx-hash but the CP confirmed deployment success")
	}
}

func (s *TestHarness) TestImageSecurity() {
    spClient, err := CreateServiceProviderTestingClient()
    s.Nil(err, "Failed to create testing client");

    var req models.DeploymentRequest
    req.TxHash = "notavalidhash"
    req.RequesterWalletAddress = "addr_test1qrrysjx7gg6e2h8qvsqc29lg37yttq6cnww72637sdgfm7c58xcwszypj5fz8mmdvkv2a7wew2tthvvftj02gdeaf4vsc849la"
    req.MaxNtx = 2
    req.Blockchain = "Cardano"
    req.ServiceType = "ml-training-cpu"
    req.Timestamp = time.Now()
	req.Params.MachineType = "cpu"
	req.Params.ModelURL = "https://gist.github.com/luigy/d63eec5cb33d9f789969fafe04ee3ae9"
	req.Params.ImageID = "registry.hub.docker.com/library/busybox"
	req.Params.RemoteNodeID = "invalidremoteid"
	req.Params.LocalNodeID = "invalidlocalid"
	req.Params.LocalPublicKey = "invalidpublickey"
	req.Constraints.Complexity = "Low"
	req.Constraints.CPU = 1500
	req.Constraints.RAM = 2000
	req.Constraints.Vram = 2000
	req.Constraints.Power = 170
	req.Constraints.Time = 1

    spClient.SendDeploymentRequest(req)

    firstUpdate, err := spClient.GetNextDeploymentUpdate()

    s.Nil(err, "Failed to get first deployment update")
    s.NotEqual(libp2p.MsgJobStatus, firstUpdate.MsgType, "Malicious SP sent an potentially malicious container but the CP started the job");

    if firstUpdate.MsgType == libp2p.MsgJobStatus {
      secondUpdate, err := spClient.GetNextDeploymentUpdate()
      s.Equal(libp2p.MsgDepResp, secondUpdate.MsgType, "We didn't recieve a response for our request");
      s.Nil(err, "Failed to get second update")
      s.Equal(false, secondUpdate.Response.Success, "We just sent a malicious container")
    }

}

func (s *TestHarness) TestPayloadSecurity() {
	spClient, err := CreateServiceProviderTestingClient()
	s.Nil(err, "Failed to create testing client");

	var req models.DeploymentRequest
	req.TxHash = "notavalidhash"
	req.RequesterWalletAddress = "addr_test1qrrysjx7gg6e2h8qvsqc29lg37yttq6cnww72637sdgfm7c58xcwszypj5fz8mmdvkv2a7wew2tthvvftj02gdeaf4vsc849la"
	req.MaxNtx = 2
	req.Blockchain = "Cardano"
	req.ServiceType = "ml-training-cpu"
	req.Timestamp = time.Now()
	req.Params.MachineType = "cpu"
	req.Params.ModelURL = "https://gist.github.com/cidkidnix/1a9245b464fc8d05e95778dc5fb6255c"
	req.Params.ImageID = "registry.gitlab.com/nunet/ml-on-gpu/ml-on-cpu-service/develop/ml-on-cpu"
	req.Params.RemoteNodeID = "invalidremoteid"
	req.Params.LocalNodeID = "invalidlocalid"
	req.Params.LocalPublicKey = "invalidpublickey"
	req.Constraints.Complexity = "Low"
	req.Constraints.CPU = 1500
	req.Constraints.RAM = 2000
	req.Constraints.Vram = 2000
	req.Constraints.Power = 170
	req.Constraints.Time = 1

	spClient.SendDeploymentRequest(req)

	firstUpdate, err := spClient.GetNextDeploymentUpdate()
	s.Nil(err, "Failed to get first deployment update")
	s.NotEqual(libp2p.MsgJobStatus, firstUpdate.MsgType, "Malicious SP send a malicious job")

	// If we have received a message a job is running, lets also confirm that the DeploymentRequest is successful as well
	if firstUpdate.MsgType == libp2p.MsgJobStatus {
		secondUpdate, err := spClient.GetNextDeploymentUpdate()
		s.Equal(libp2p.MsgDepResp, secondUpdate.MsgType, "We didn't receive a DeploymentResponse for our DeploymentRequest")
		s.Nil(err, "Failed to get second deployment update")
		s.Equal(false, secondUpdate.Response.Success, "Malicious SP sent a malicious job but the CP confirmed deployment success")
	}
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
	db.ConnectDatabase()

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
	pair, err := GenerateTestKeyPair();

	runAsServer := true;
	libp2p.RunNode(pair, runAsServer)
	return err
}

// Interface for SP to communicate with testing CP
type SPTestClient struct
{
	reader *bufio.Reader
	writer *bufio.Writer
}

// Convenience type for DeploymentUpdate with unmarshalled data
type CPUpdate struct
{
	MsgType string
	Response models.DeploymentResponse
	Services models.Services
}

func CreateServiceProviderTestingClient() (SPTestClient, error) {
	ctx := context.Background()

	SetupDMSTestingConfiguration("client", 9000)

	pair, err := GenerateTestKeyPair()
	host, dht, err := libp2p.NewHost(ctx, pair, false)
	p2p := *libp2p.DMSp2pInit(host, dht)
	err = p2p.BootstrapNode(ctx)

	cpID := GetTestComputeProviderID()
	stream, err := host.NewStream(context.Background(), cpID, libp2p.DepReqProtocolID)

	reader := bufio.NewReader(stream)
	writer := bufio.NewWriter(stream)

	return SPTestClient{reader,writer}, err
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
