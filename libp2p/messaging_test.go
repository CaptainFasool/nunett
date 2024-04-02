package libp2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/models"
)

func TestChat(t *testing.T) {
	ctx := context.Background()

	// set port to 0 to use available port
	config.SetConfig("p2p.listen_address", []string{
		"/ip4/0.0.0.0/tcp/0",
		"/ip4/0.0.0.0/udp/0/quic",
	})

	// initialize first node
	defer ctx.Done()
	priv1, _, _ := GenerateKey(time.Now().UnixNano())

	host1, _, err := NewHost(ctx, priv1, true)

	if err != nil {
		t.Fatalf("First Node Initialization Failed: %v", err)
	}
	defer host1.Close()

	host1.SetStreamHandler(protocol.ID(ChatProtocolID), chatStreamHandler)

	// 0 chat requests expected now
	incomingReq, err := IncomingChatRequests()
	assert.Nil(t, incomingReq)
	assert.Equal(t, "no incoming message stream", err.Error())

	time.Sleep(10 * time.Millisecond)
	// initialize second node
	priv2, _, _ := GenerateKey(time.Now().UnixNano())

	host2, _, err := NewHost(ctx, priv2, true)
	if err != nil {
		t.Fatalf("Second Node Initialization Failed: %v", err)
	}
	defer host2.Close()

	host2.Peerstore().AddAddrs(host1.ID(), host1.Addrs(), peerstore.PermanentAddrTTL)
	host2.Peerstore().AddPubKey(host1.ID(), host1.Peerstore().PubKey(host1.ID()))

	if err := host2.Connect(ctx, peer.AddrInfo{ID: host1.ID(), Addrs: host1.Addrs()}); err != nil {
		t.Errorf("Unable to connect ---- %v ", err)
	}

	time.Sleep(1 * time.Second)

	// try to connect 5 times
	attempt := 0
	for host2.Network().Connectedness(host1.ID()).String() != "Connected" && attempt < 5 {
		if err := host2.Connect(ctx, host1.Peerstore().PeerInfo(host1.ID())); err != nil {
			t.Errorf("Unable to connect - %v ", err)
		}
		time.Sleep(1 * time.Second)
		attempt++
	}

	stream, err := host2.NewStream(ctx, host1.ID(), protocol.ID(ChatProtocolID))
	if err != nil {
		t.Fatalf("Error Creating Stream From Second Node to First Node: %v", err)
	}

	w := bufio.NewWriter(stream)
	writeString(w, "Hello, just testing\n")
	w.Flush()

	time.Sleep(time.Second)

	// 1 chat requests expected now
	incomingReq, err = IncomingChatRequests()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(incomingReq))

	r := bufio.NewReader(inboundChatStreams[0])
	msg, err := readString(r)
	if err != nil {
		t.Fatalf("Couldn't Read from Stream: %v", err)
	}

	assert.Equal(t, "Hello, just testing\n", msg)

}

func TestWrongFormatDepReq(t *testing.T) {
	ctx := context.Background()

	// set port to 0 to use available port
	config.SetConfig("p2p.listen_address", []string{
		"/ip4/0.0.0.0/tcp/0",
		"/ip4/0.0.0.0/udp/0/quic",
	})

	// initialize first node
	defer ctx.Done()
	priv1, _, _ := GenerateKey(time.Now().UnixNano())

	host1, _, err := NewHost(ctx, priv1, true)

	if err != nil {
		t.Fatalf("First Node Initialization Failed: %v", err)
	}
	defer host1.Close()

	host1.SetStreamHandler(protocol.ID(DepReqProtocolID), depReqStreamHandler)

	time.Sleep(10 * time.Millisecond)

	// initialize second node
	priv2, _, _ := GenerateKey(time.Now().UnixNano())

	host2, _, err := NewHost(ctx, priv2, true)
	if err != nil {
		t.Fatalf("Second Node Initialization Failed: %v", err)
	}
	defer host2.Close()

	host2.Peerstore().AddAddrs(host1.ID(), host1.Addrs(), peerstore.PermanentAddrTTL)
	host2.Peerstore().AddPubKey(host1.ID(), host1.Peerstore().PubKey(host1.ID()))

	if err := host2.Connect(ctx, peer.AddrInfo{ID: host1.ID(), Addrs: host1.Addrs()}); err != nil {
		t.Errorf("Unable to connect ---- %v ", err)
	}

	time.Sleep(1 * time.Second)

	// try to connect 5 times
	attempt := 0
	for host2.Network().Connectedness(host1.ID()).String() != "Connected" && attempt < 5 {
		if err := host2.Connect(ctx, host1.Peerstore().PeerInfo(host1.ID())); err != nil {
			t.Errorf("Unable to connect - %v ", err)
		}
		time.Sleep(1 * time.Second)
		attempt++
	}

	stream, err := host2.NewStream(ctx, host1.ID(), protocol.ID(DepReqProtocolID))
	if err != nil {
		t.Fatalf("Error Creating Stream From Second Node to First Node: %v", err)
	}

	fmt.Println("Stream Created: ", stream)

	wrongDepReq := "hi"

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	writeString(rw.Writer, fmt.Sprintf("%v\n", wrongDepReq))
	rw.Flush()

	time.Sleep(2 * time.Second)

	msg, err := readString(rw.Reader)
	if err != nil {
		t.Fatalf("Couldn't Read from Stream: %v", err)
	}

	depUpdate := models.DeploymentUpdate{}
	json.Unmarshal([]byte(msg), &depUpdate)
	assert.Equal(t, "DepResp", depUpdate.MsgType)

	depResp := models.DeploymentResponse{}
	json.Unmarshal([]byte(depUpdate.Msg), &depResp)

	assert.False(t, depResp.Success)
	assert.Equal(t, "Unable to decode deployment request", depResp.Content)
}
