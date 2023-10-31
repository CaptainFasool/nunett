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
	"log"

	"github.com/stretchr/testify/suite"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/tracing"
	"gitlab.com/nunet/device-management-service/internal/messaging"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/crypto"
)

// This is essentially MyTestSuite defined in cli_test.go and gets pulled in
// TODO put this in a common place and re-use
type TheTestSuite struct {
	suite.Suite
	sync.WaitGroup
	cleanup func(context.Context) error
}

func TestTheTestSuite(t *testing.T) {
	s := new(TheTestSuite)
	suite.Run(t, s)
}

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

func (s *TheTestSuite) SetupSuite() {
	os.Mkdir("/tmp/nunet.test", 0755)
	ioutil.WriteFile("/tmp/nunet.test/metadataV2.json", []byte(testMetadata), 0644);
	config.LoadConfig()
	addrs := [2]string{"/ip4/0.0.0.0/tcp/9123", "/ip4/0.0.0.0/udp/9123/quic"}
	config.SetConfig("general.metadata_path", "/tmp/nunet.test")
	config.SetConfig("p2p.listen_address", addrs)
	db.ConnectDatabase()

	go messaging.DeploymentWorker()

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

	db_res := db.DB.Create(&availableResources)
	s.NotNil(db_res, "Failed to create available resources")

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

	db_res = db.DB.Create(&freeResources)
	s.NotNil(db_res, "Failed to create free resources")

	s.cleanup = tracing.InitTracer()
	s.WaitGroup.Add(1)
	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519, // Select your key type. Ed25519 are nice short
		-1,             // Select key length when possible (i.e. RSA).
	)
	s.Nil(err, "Failed to generate private key");

	libp2p.RunNode(priv, true);
}

func RunTestNode() (host.Host) {
	os.Mkdir("/tmp/nunet.test2", 0755)
	config.LoadConfig()
	ioutil.WriteFile("/tmp/nunet.test2/metadataV2.json", []byte(testMetadata), 0644);
	config.SetConfig("general.metadata_path", "/tmp/nunet.test2")

	ctx := context.Background()

	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519, // Select your key type. Ed25519 are nice short
		-1,             // Select key length when possible (i.e. RSA).
	)

	host, dht, err := libp2p.NewHost(ctx, priv, false)
	if err != nil {
		panic(err)
	}

	p2p := *libp2p.DMSp2pInit(host, dht)

	err = p2p.BootstrapNode(ctx)

	if err != nil {
	}

	return host
}

func (s *TheTestSuite) TearDownSuite() {
	s.WaitGroup.Done()
	s.cleanup(context.Background())
	os.RemoveAll("/tmp/nunet.test")
}

func (s *TheTestSuite) TestWIP() {
	h := RunTestNode()

	p2p := libp2p.GetP2P()
	dmsid := p2p.Host.ID()

	stream, err := h.NewStream(context.Background(), dmsid, libp2p.DepReqProtocolID)
	s.Nil(err, "Failed to connect to node ", err);

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

	msg, json_err := json.Marshal(req)
	s.Nil(json_err, "Failed to marshal request to json")

	writer := bufio.NewWriter(stream)

	log.Println("Sending job")
	_, write_err := writer.WriteString(fmt.Sprintf("%s\n", msg))
	s.Nil(write_err, "Failed to write to p2p stream")
	flush_err := writer.Flush()
	s.Nil(flush_err, "Failed to flush")

	reader := bufio.NewReader(stream)

	log.Println("Waiting for response")

	// The first response is just the JobStatus, which should be sufficient for proving things out
	// but we will wait for the DeploymentResponse to get an affirmed "success" value...
	response_msg, read_err := reader.ReadString('\n')
	s.Nil(read_err, "Failed to read from p2p stream")

	response_msg, read_err = reader.ReadString('\n')
	s.Nil(read_err, "Failed or something to read from p2p stream")

	// This is the second deployment update which contains the status
	res := models.DeploymentUpdate{}
	json_err = json.Unmarshal([]byte(response_msg), &res)
	s.Nil(json_err, "Failed to unmarshal response")

	s.Equal(res.MsgType, "DepResp");
	resp := models.DeploymentResponse{}
	json_err = json.Unmarshal([]byte(res.Msg), &resp)

	s.Equal(false, resp.Success, "Invalid tx-hash sent but job ran successfully")
}
