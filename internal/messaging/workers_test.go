package messaging

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/config"
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDepReq(t *testing.T) {
	// start depreq worker
	go DeploymentWorker()

	// mock db
	mockDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Errorf("error trying to initialize mock db: %v", err)
	}
	db.DB = mockDB

	// set host listening address to tcp only and to use available port
	config.SetConfig("p2p.listen_address", []string{
		"/ip4/0.0.0.0/tcp/0",
	})

	// use a random channel name
	channelName := utils.RandomString(10)

	ctx := context.Background()

	// initialize first node
	defer ctx.Done()
	priv1, _, _ := libp2p.GenerateKey(time.Now().Unix())
	var metadata models.MetadataV2
	metadata.AllowCardano = false
	metadata.Network = channelName
	meta, _ := json.Marshal(metadata)
	libp2p.FS = afero.NewMemMapFs()
	libp2p.AFS = &afero.Afero{Fs: libp2p.FS}
	// create test files and directories
	libp2p.AFS.MkdirAll("/etc/nunet", 0755)
	afero.WriteFile(libp2p.AFS, "/etc/nunet/metadataV2.json", meta, 0644)

	libp2p.RunNode(priv1, true, true)

	p2p := libp2p.GetP2P()

	p2p.Host.SetStreamHandler(libp2p.DepReqProtocolID, mockDepReqStreamHandler)

	// initialize second node
	priv2, _, _ := libp2p.GenerateKey(time.Now().Unix())
	host2, idht2, err := libp2p.NewHost(ctx, priv2, true)
	if err != nil {
		t.Fatalf("Second Node Initialization Failed: %v", err)
	}
	defer host2.Close()

	err = libp2p.Bootstrap(ctx, host2, idht2)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}

	host2.Peerstore().AddAddrs(p2p.Host.ID(), p2p.Host.Addrs(), peerstore.PermanentAddrTTL)
	host2.Peerstore().AddPubKey(p2p.Host.ID(), p2p.Host.Peerstore().PubKey(p2p.Host.ID()))

	if err := host2.Connect(ctx, p2p.Host.Peerstore().PeerInfo(p2p.Host.ID())); err != nil {
		t.Errorf("Unable to connect - %v ", err)
	}
	time.Sleep(3 * time.Second)

	for host2.Network().Connectedness(p2p.Host.ID()).String() != "Connected" {
		if err := host2.Connect(ctx, p2p.Host.Peerstore().PeerInfo(p2p.Host.ID())); err != nil {
			t.Errorf("Unable to connect to host 1 - %v ", err)
		}
		fmt.Println("Host 2 to Host 1:", host2.Network().Connectedness(p2p.Host.ID()).String())
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

	var depUpdate models.DeploymentUpdate
	err = json.Unmarshal([]byte(msg), &depUpdate)
	if err != nil {
		t.Error("Unable to Parse Deployment Update", err)
	}

	assert.Equal(t, "DepResp", depUpdate.MsgType)

	var depResp models.DeploymentResponse
	err = json.Unmarshal([]byte(depUpdate.Msg), &depResp)
	if err != nil {
		t.Error("Unable to Parse Deployment Response", err)
	}
	assert.False(t, depResp.Success)
	assert.Equal(t, "Unknown service type.", depResp.Content)

	correctDepReq := models.DeploymentRequest{
		RequesterWalletAddress: host2.ID().Pretty(),
		MaxNtx:                 10,
		Blockchain:             "cardano",
		ServiceType:            "cardano_node",
		Timestamp:              time.Now().In(time.UTC),
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

	err = json.Unmarshal([]byte(msg), &depUpdate)
	if err != nil {
		t.Error("Unable to Parse Deployment Response", err)
	}

	err = json.Unmarshal([]byte(depUpdate.Msg), &depResp)
	if err != nil {
		t.Error("Unable to Parse Deployment Response", err)
	}
	assert.False(t, depResp.Success)
	assert.Equal(t, "Cardano Node Deployment Failed. Unable to download "+utils.KernelFileURL, depResp.Content)
}

func mockDepReqStreamHandler(stream network.Stream) {
	ctx := context.Background()
	defer ctx.Done()
	zlog.InfoContext(ctx, "Got a new test depReq stream!")

	r := bufio.NewReader(stream)
	str, err := r.ReadString('\n')
	if err != nil {
		zlog.Sugar().Errorf("failed to read from new stream buffer - %v", err)
		w := bufio.NewWriter(stream)

		_, err := w.WriteString(fmt.Sprintf("%s\n", "Unable to read DepReq. Closing Stream."))
		if err != nil {
			zlog.Sugar().Errorf("failed to write to stream after %s - %v", "unable to read depReq", err)
		}

		err = w.Flush()
		if err != nil {
			zlog.Sugar().Errorf("failed to flush stream after %s - %v", "unable to read depReq", err)
		}

		err = stream.Close()
		if err != nil {
			zlog.Sugar().Errorf("failed to close stream after %s - %v", "unable to read depReq", err)
		}
		return
	}

	zlog.Sugar().DebugfContext(ctx, "[depReq recv] message: %s", str)

	libp2p.InboundDepReqStream = stream

	depreqMessage := models.DeploymentRequest{}
	err = json.Unmarshal([]byte(str), &depreqMessage)
	if err != nil {
		zlog.ErrorContext(ctx, fmt.Sprintf("unable to decode deployment request: %v", err))
		// XXX : might be best to propagate context through depReq/depResp to encompass everything done starting with a single depReq
		depRes := models.DeploymentResponse{Success: false, Content: "Unable to decode deployment request"}
		depResBytes, _ := json.Marshal(depRes)
		libp2p.DeploymentUpdate(libp2p.MsgDepResp, string(depResBytes), true)
	} else {

		libp2p.DepReqQueue <- depreqMessage

	}
}
