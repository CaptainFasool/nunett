package libp2p

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/utils"
)

func TestChat(t *testing.T) {
	ctx := context.Background()

	// initialize first node
	defer ctx.Done()
	priv1, _, _ := GenerateKey(time.Now().Unix())

	host1, idht1, err := NewHost(ctx, priv1, true)
	// p2p = *DMSp2pInit(host1, idht1)

	if err != nil {
		t.Fatalf("First Node Initialization Failed: %v", err)
	}
	defer host1.Close()

	host1.SetStreamHandler(protocol.ID(ChatProtocolID), ChatStreamHandler)

	err = Bootstrap(ctx, host1, idht1)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}

	// 0 chat requests expected now
	incomingReq, err := incomingChatRequests()
	assert.Nil(t, incomingReq)
	assert.Equal(t, "No Incoming Message Stream.", err.Error())

	// initialize second node
	priv2, _, _ := GenerateKey(time.Now().Unix())
	host2, idht2, err := NewHost(ctx, priv2, true)
	if err != nil {
		t.Fatalf("Second Node Initialization Failed: %v", err)
	}
	defer host2.Close()

	err = Bootstrap(ctx, host2, idht2)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}

	go Discover(context.Background(), P2P{
		Host: host1,
		DHT:  idht1,
	}, CIRendevousPoint)
	go Discover(ctx, P2P{
		Host: host2,
		DHT:  idht2,
	}, CIRendevousPoint)

	host2.Peerstore().AddAddrs(host1.ID(), host1.Addrs(), peerstore.PermanentAddrTTL)
	host2.Peerstore().AddPubKey(host1.ID(), host1.Peerstore().PubKey(host1.ID()))

	if err := host2.Connect(ctx, host1.Peerstore().PeerInfo(host1.ID())); err != nil {
		t.Errorf("Unable to connect ---- %v ", err)
	}

	time.Sleep(1 * time.Second)

	for host2.Network().Connectedness(host1.ID()).String() != "Connected" {
		if err := host2.Connect(ctx, host1.Peerstore().PeerInfo(host1.ID())); err != nil {
			t.Errorf("Unable to connect - %v ", err)
		}
		time.Sleep(1 * time.Second)

	}

	stream, err := host2.NewStream(ctx, host1.ID(), protocol.ID(ChatProtocolID))
	if err != nil {
		t.Fatalf("Error Creating Stream From Second Node to First Node: %v", err)
	}

	w := bufio.NewWriter(stream)
	writeData(w, "Hello, just testing\n")
	w.Flush()

	time.Sleep(time.Second)

	// 1 chat requests expected now
	incomingReq, err = incomingChatRequests()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(incomingReq))

	r := bufio.NewReader(inboundChatStreams[0])
	msg, err := readData(r)
	if err != nil {
		t.Fatalf("Couldn't Read from Stream: %v", err)
	}

	assert.Equal(t, "Hello, just testing\n", msg)

}

func TestWrongFormatDepReq(t *testing.T) {
	ctx := context.Background()

	// initialize first node
	defer ctx.Done()
	priv1, _, _ := GenerateKey(time.Now().Unix())

	host1, idht1, err := NewHost(ctx, priv1, true)

	if err != nil {
		t.Fatalf("First Node Initialization Failed: %v", err)
	}
	defer host1.Close()

	host1.SetStreamHandler(protocol.ID(DepReqProtocolID), DepReqStreamHandler)

	err = Bootstrap(ctx, host1, idht1)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}

	// initialize second node
	priv2, _, _ := GenerateKey(time.Now().Unix())
	host2, idht2, err := NewHost(ctx, priv2, true)
	if err != nil {
		t.Fatalf("Second Node Initialization Failed: %v", err)
	}
	defer host2.Close()

	err = Bootstrap(ctx, host2, idht2)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	testRendezvous := utils.RandomString(20)

	go Discover(ctx, P2P{
		Host: host1,
		DHT:  idht1,
	}, testRendezvous)
	go Discover(ctx, P2P{
		Host: host2,
		DHT:  idht2,
	}, testRendezvous)

	host2.Peerstore().AddAddrs(host1.ID(), host1.Addrs(), peerstore.PermanentAddrTTL)
	host2.Peerstore().AddPubKey(host1.ID(), host1.Peerstore().PubKey(host1.ID()))

	if err := host2.Connect(ctx, host1.Peerstore().PeerInfo(host1.ID())); err != nil {
		t.Errorf("Unable to connect ---- %v ", err)
	}

	time.Sleep(1 * time.Second)

	for host2.Network().Connectedness(host1.ID()).String() != "Connected" {
		if err := host2.Connect(ctx, host1.Peerstore().PeerInfo(host1.ID())); err != nil {
			t.Errorf("Unable to connect - %v ", err)
		}
		time.Sleep(1 * time.Second)

	}

	stream, err := host2.NewStream(ctx, host1.ID(), protocol.ID(DepReqProtocolID))
	if err != nil {
		t.Fatalf("Error Creating Stream From Second Node to First Node: %v", err)
	}

	wrongDepReq := "hi"

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	writeData(rw.Writer, fmt.Sprintf("%v\n", wrongDepReq))
	rw.Flush()

	time.Sleep(2 * time.Second)

	msg, err := readData(rw.Reader)
	if err != nil {
		t.Fatalf("Couldn't Read from Stream: %v", err)
	}

	assert.Equal(t, "Unable to decode deployment request\n", msg)
}
