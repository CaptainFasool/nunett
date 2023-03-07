package messaging

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
)

func TestDepReq(t *testing.T) {
	// start depreq worker
	go DeploymentWorker()

	ctx := context.Background()

	// initialize first node
	defer ctx.Done()
	priv1, _, _ := libp2p.GenerateKey(time.Now().Unix())
	var metadata models.MetadataV2
	metadata.AllowCardano = false
	meta, _ := json.Marshal(metadata)
	libp2p.FS = afero.NewMemMapFs()
	libp2p.AFS = &afero.Afero{Fs: libp2p.FS}
	// create test files and directories
	libp2p.AFS.MkdirAll("/etc/nunet", 0755)
	afero.WriteFile(libp2p.AFS, "/etc/nunet/metadataV2.json", meta, 0644)

	libp2p.RunNode(priv1)

	p2p := libp2p.GetP2P()

	// initialize second node
	priv2, _, _ := libp2p.GenerateKey(time.Now().Unix())
	host2, idht2, err := libp2p.NewHost(ctx, 9501, priv2)
	if err != nil {
		t.Fatalf("Second Node Initialization Failed: %v", err)
	}
	defer host2.Close()

	err = libp2p.Bootstrap(ctx, host2, idht2)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	testRendezvous := utils.RandomString(10)

	go libp2p.Discover(context.Background(), p2p.Host, p2p.DHT, testRendezvous)
	go libp2p.Discover(ctx, host2, idht2, testRendezvous)

	host2.Peerstore().AddAddrs(p2p.Host.ID(), p2p.Host.Addrs(), peerstore.PermanentAddrTTL)
	host2.Peerstore().AddPubKey(p2p.Host.ID(), p2p.Host.Peerstore().PubKey(p2p.Host.ID()))

	if err := host2.Connect(ctx, p2p.Host.Peerstore().PeerInfo(p2p.Host.ID())); err != nil {
		t.Errorf("Unable to connect - %v ", err)
	}
	time.Sleep(3 * time.Second)

	for host2.Network().Connectedness(p2p.Host.ID()).String() != "Connected" {
		if err := host2.Connect(ctx, p2p.Host.Peerstore().PeerInfo(p2p.Host.ID())); err != nil {
			t.Errorf("Unable to connect - %v ", err)
		}
		time.Sleep(1 * time.Second)

	}

	stream, err := host2.NewStream(ctx, p2p.Host.ID(), protocol.ID(libp2p.DepReqProtocolID))
	if err != nil {
		t.Skipf("Error Creating Stream From Second Node to First Node: %v", err)
		t.Skipped()
	}

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	wrongDepReqServType := `{"address_user": "foobar-address", "service_type": "BAD_SERVICE_TYPE_FOR_TEST"}`

	rw.WriteString(fmt.Sprintf("%v\n", wrongDepReqServType))
	rw.Flush()

	time.Sleep(2 * time.Second)

	msg, err := rw.ReadString('\n')
	if err != nil {
		t.Fatalf("Couldn't Read from Stream: %v", err)
	}

	var resp models.DeploymentResponse
	err = json.Unmarshal([]byte(msg), &resp)
	if err != nil {
		t.Error("Unable to Parse Deployment Response", err)
	}

	assert.False(t, resp.Success)
	assert.Equal(t, "Unknown service type.", resp.Content)

	correctDepReq := models.DeploymentRequest{
		AddressUser: host2.ID().Pretty(),
		MaxNtx:      10,
		Blockchain:  "cardano",
		ServiceType: "cardano_node",
		Timestamp:   time.Now(),
	}
	correctJsonDepReq, err := json.Marshal(correctDepReq)
	err = stream.Close()
	if err != nil {
		t.Error("Error Closing Frist Stream:", err.Error())
	}

	utils.KernelFileURL = "https://" + utils.RandomString(20) + ".com/vmlinux"

	stream, err = host2.NewStream(ctx, p2p.Host.ID(), protocol.ID(libp2p.DepReqProtocolID))
	if err != nil {
		t.Fatalf("Error Creating Second Stream From Second Node to First Node: %v", err)
	}

	rw = bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	rw.WriteString(fmt.Sprintf("%v\n", string(correctJsonDepReq)))
	rw.Flush()

	time.Sleep(2 * time.Second)

	msg, err = rw.ReadString('\n')
	if err != nil {
		t.Fatalf("Couldn't Read from Stream: %v", err)
	}

	err = json.Unmarshal([]byte(msg), &resp)
	if err != nil {
		t.Error("Unable to Parse Deployment Response", err)
	}

	assert.False(t, resp.Success)
	assert.Equal(t, "Cardano Node Deployment Failed. Unable to download "+utils.KernelFileURL, resp.Content)
}
