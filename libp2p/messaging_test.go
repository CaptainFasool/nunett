package libp2p

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/assert"
	"gitlab.com/nunet/device-management-service/utils"
)

func TestChat(t *testing.T) {
	ctx := context.Background()

	// initialize first node
	defer ctx.Done()
	priv1, _, _ := GenerateKey(time.Now().Unix())

	host1, idht1, err := NewHost(ctx, 9500, priv1)
	// p2p = *DMSp2pInit(host1, idht1)

	if err != nil {
		t.Fatalf("First Node Initialization Failed: %v", err)
	}
	defer host1.Close()

	host1.SetStreamHandler(protocol.ID(ChatProtocolID), chatStreamHandler)

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
	host2, idht2, err := NewHost(ctx, 9501, priv2)
	if err != nil {
		t.Fatalf("Second Node Initialization Failed: %v", err)
	}
	defer host2.Close()

	err = Bootstrap(ctx, host2, idht2)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}

	go Discover(ctx, host1, idht1, CIRendevousPoint)
	go Discover(ctx, host2, idht2, CIRendevousPoint)

	if err := host2.Connect(context.Background(), host1.Peerstore().PeerInfo(host1.ID())); err != nil {
		t.Errorf("Unable to connect ---- %v ", err)
	}

	connectedness := host2.Network().Connectedness(host1.ID())
	if connectedness.String() != "Connected" {
		t.Log("Unable to Proceed - Hosts Not Connected")
		t.Skip("Unable to Proceed - Hosts Not Connected")
		t.Skipped()
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

	host1, idht1, err := NewHost(ctx, 9500, priv1)

	if err != nil {
		t.Fatalf("First Node Initialization Failed: %v", err)
	}
	defer host1.Close()

	host1.SetStreamHandler(protocol.ID(DepReqProtocolID), depReqStreamHandler)

	err = Bootstrap(ctx, host1, idht1)
	if err != nil {
		t.Fatalf("Bootstrap returned error: %v", err)
	}

	// initialize second node
	priv2, _, _ := GenerateKey(time.Now().Unix())
	host2, idht2, err := NewHost(ctx, 9501, priv2)
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

	go Discover(ctx, host1, idht1, testRendezvous)
	go Discover(ctx, host2, idht2, testRendezvous)

	if err := host2.Connect(context.Background(), host1.Peerstore().PeerInfo(host1.ID())); err != nil {
		t.Log("Unable to Proceed - Hosts Not Connected")
		t.Skip("Unable to Proceed - Hosts Not Connected")
		t.Skipped()
	}

	connectedness := host2.Network().Connectedness(host1.ID())
	if connectedness.String() != "Connected" {
		t.Log("Unable to Proceed - Hosts Not Connected")
		t.Skip("Unable to Proceed - Hosts Not Connected")
		t.Skipped()
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
