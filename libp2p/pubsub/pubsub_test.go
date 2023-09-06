package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"gitlab.com/nunet/device-management-service/libp2p"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pPS "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/stretchr/testify/suite"
)

type PubSubTestSuite struct {
	suite.Suite
	ctx          context.Context
	node1        *psNodeConfigTest
	node2        *psNodeConfigTest
	observerNode *psNodeConfigTest
	topicName    string
	wg           *sync.WaitGroup
}

type psNodeConfigTest struct {
	host   host.Host
	idht   *dht.IpfsDHT
	pubsub *PubSubPeer
	topic  *PsTopicSubscription
	msg    string
}

func startupHostPubSubTest(ctx context.Context, peers ...host.Host) (*psNodeConfigTest, error) {
	var err error

	zlog.Sugar().Debug("Creating host")
	priv, _, _ := libp2p.GenerateKey(0)
	host, idht, err := libp2p.NewHost(ctx, priv, true)
	if err != nil {
		return nil, err
	}

	if err := connectToPeers(ctx, host, peers...); err != nil {
		return nil, err
	}

	ps, err := newTestGossipPubSub(ctx, host)
	if err != nil {
		return nil, err
	}

	return &psNodeConfigTest{
		host:   host,
		idht:   idht,
		pubsub: ps,
	}, nil
}

func (s *PubSubTestSuite) SetupSuite() {
	var err error
	s.ctx = context.Background()

	s.observerNode, err = startupHostPubSubTest(s.ctx)
	s.Require().NoError(err)

	s.node1, err = startupHostPubSubTest(s.ctx, s.observerNode.host)
	s.Require().NoError(err)

	s.node2, err = startupHostPubSubTest(s.ctx, s.observerNode.host)
	s.Require().NoError(err)

	s.topicName = "topicTest"
	s.node1.msg = "I love lasagna"
	s.node2.msg = "me too"

	s.wg = &sync.WaitGroup{}
}

// newTestGossipPubSub creates a new GossipSub instance for tests
func newTestGossipPubSub(ctx context.Context, host host.Host) (*PubSubPeer, error) {
	gs, err := libp2pPS.NewGossipSub(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("Failed to create gossipsub, Error: %w", err)
	}
	return &PubSubPeer{gs}, nil
}

func (s *PubSubTestSuite) TearDownSuite() {
	s.NoError(s.observerNode.host.Close())
	s.NoError(s.node1.host.Close())
	s.NoError(s.node2.host.Close())
}

// TestPubSub tests the libp2p's GossipSub implementation and message publishing
// and retrieving processes
func (s *PubSubTestSuite) TestPubSub() {
	// Equal strings for node1 and observerNode
	var err error

	s.observerNode.topic, err = s.observerNode.pubsub.JoinTopic(s.topicName)
	s.Require().NoError(err)

	s.node1.topic, err = s.node1.pubsub.JoinTopic(s.topicName)
	s.Require().NoError(err)

	s.node2.topic, err = s.node2.pubsub.JoinTopic(s.topicName)
	s.Require().NoError(err)

	msgCh := make(chan *libp2pPS.Message)
	go s.observerNode.topic.listenForMessages(s.ctx, msgCh)
	time.Sleep(4 * time.Second)

	// Check if we observerNode received the right messages from the right peers
	// Just to messages were send so we iterate 2 times

	var m string
	var msg *libp2pPS.Message

	s.NoError(s.node1.topic.Publish(s.node1.msg))
	msg = <-msgCh
	err = json.Unmarshal(msg.Data, &m)
	if err != nil {
		s.Require().NoError(err)
	}
	s.Equal(msg.ReceivedFrom, s.node1.host.ID())
	s.Equal(s.node1.msg, m)

	time.Sleep(4 * time.Second)
	s.NoError(s.node2.topic.Publish(s.node2.msg))
	msg = <-msgCh
	err = json.Unmarshal(msg.Data, &m)
	if err != nil {
		s.Require().NoError(err)
	}
	s.Equal(msg.ReceivedFrom, s.node2.host.ID())
	s.Equal(s.node2.msg, m)

	zlog.Sugar().Debug("Finalizing")

	// check if we received only two messages, one for each node
	// check if there are more messages

	s.ctx.Deadline()
}

func TestPubSubSuite(t *testing.T) {
	suite.Run(t, new(PubSubTestSuite))
}

func connectToPeers(ctx context.Context, host host.Host, peers ...host.Host) error {
	for _, p := range peers {
		if err := host.Connect(ctx, p.Peerstore().PeerInfo(p.ID())); err != nil {
			return err
		}
	}
	return nil
}
