package pubsub

import (
	"context"
	"fmt"
	"testing"
	"time"

	libp2pPS "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/stretchr/testify/suite"
	"gitlab.com/nunet/device-management-service/libp2p"
)

type PubSubTestSuite struct {
	suite.Suite
	ctx          context.Context
	node1        psNodeConfigTest
	node2        psNodeConfigTest
	observerNode psNodeConfigTest
	topicName    string
}

type psNodeConfigTest struct {
	host   host.Host
	pubsub *PubSubPeer
	topic  *PsTopicSubscription
	msg    string
}

func (s *PubSubTestSuite) SetupSuite() {
	var err error
	s.ctx = context.Background()

	// create observerNode host and pubsub instance
	priv3, _, _ := libp2p.GenerateKey(time.Now().Unix())
	s.observerNode.host, _, err = libp2p.NewHost(s.ctx, priv3, true)
	s.Require().NoError(err)

	s.observerNode.pubsub, err = newTestGossipPubSub(s.ctx, s.observerNode.host)
	s.Require().NoError(err)

	// create node1 host and pubsub instance
	priv1, _, _ := libp2p.GenerateKey(time.Now().Unix())
	s.node1.host, _, err = libp2p.NewHost(s.ctx, priv1, true)
	s.Require().NoError(err)

	s.node1.pubsub, err = newTestGossipPubSub(s.ctx, s.node1.host)
	s.Require().NoError(err)

	// create node2 host and pubsub instance
	priv2, _, _ := libp2p.GenerateKey(time.Now().Unix())
	s.node2.host, _, err = libp2p.NewHost(s.ctx, priv2, true)
	s.Require().NoError(err)

	s.node2.pubsub, err = newTestGossipPubSub(s.ctx, s.node2.host)
	s.Require().NoError(err)

	s.topicName = "topicTest"
	s.node1.msg = "I love lasagna"
	s.node2.msg = "me too"
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

	s.NoError(s.node1.topic.Publish(s.node1.msg))

	s.NoError(s.node2.topic.Publish(s.node2.msg))

	// Check if we observerNode received the right messages from the right peers
	// Just to messages were send so we iterate 2 times
	for i := 0; i < 2; i++ {
		msg := <-msgCh
		if msg.ReceivedFrom == s.node1.host.ID() {
			s.Equal(s.node1.msg, msg.Message.String())

		} else if msg.ReceivedFrom == s.node2.host.ID() {
			s.Equal(s.node2.msg, msg.Message.String())
		}
	}
	// check if we received only two messages, one for each node
	// check if there are more messages

	s.ctx.Deadline()
}

func TestPubSubSuite(t *testing.T) {
	suite.Run(t, new(PubSubTestSuite))
}
