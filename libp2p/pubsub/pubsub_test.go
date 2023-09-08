package pubsub

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	libp2pPS "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/stretchr/testify/suite"
)

type PubSubTestSuite struct {
	suite.Suite
	ctx          context.Context
	nodeAB       *psNodeConfigTest
	nodeXY       *psNodeConfigTest
	observerNode *psNodeConfigTest
	topicName    string
}

type psNodeConfigTest struct {
	host     host.Host
	pubsub   *PubSub
	topicSub *PsTopicSubscription
	msg      string
}

func (s *PubSubTestSuite) SetupSuite() {
	var err error
	s.ctx = context.Background()

	s.observerNode, err = startupHostPubSubTest(context.Background())
	s.Require().NoError(err)

	s.nodeAB, err = startupHostPubSubTest(context.Background(), s.observerNode.host)
	s.Require().NoError(err)

	s.nodeXY, err = startupHostPubSubTest(context.Background(), s.observerNode.host)
	s.Require().NoError(err)

	s.topicName = "test topic"
	s.nodeAB.msg = "I love lasagna!"
	s.nodeXY.msg = "Me too!"
	s.observerNode.msg = "I must not catch my own message"
}

func (s *PubSubTestSuite) TearDownSuite() {
	s.NoError(s.observerNode.host.Close())
	s.NoError(s.nodeAB.host.Close())
	s.NoError(s.nodeXY.host.Close())
}

// TestPubSub tests the libp2p's GossipSub implementation and message publishing
// and retrieving processes
func (s *PubSubTestSuite) TestPubSub() {
	var err error

	// Joining observerNode to the topic and making it listen to all coming messages
	s.observerNode.topicSub, err = s.observerNode.pubsub.JoinSubscribeTopic(s.topicName)
	s.Require().NoError(err)
	msgCh := make(chan *libp2pPS.Message)
	go s.observerNode.topicSub.ListenForMessages(context.Background(), msgCh)
	time.Sleep(1 * time.Second)

	// joining nodeAB into the topic
	s.nodeAB.topicSub, err = s.nodeAB.pubsub.JoinSubscribeTopic(s.topicName)
	s.Require().NoError(err)

	// joining nodeXY into the topic
	s.nodeXY.topicSub, err = s.nodeXY.pubsub.JoinSubscribeTopic(s.topicName)
	s.Require().NoError(err)

	// Publishing message with observerNode which should be ignored by
	// its own ListenForMessages()
	s.NoError(s.observerNode.topicSub.Publish(s.observerNode.msg))
	time.Sleep(1 * time.Second)

	// Publishing message with nodeAB and nodeXY
	s.NoError(s.nodeAB.topicSub.Publish(s.nodeAB.msg))
	time.Sleep(1 * time.Second)
	s.NoError(s.nodeXY.topicSub.Publish(s.nodeXY.msg))

	// Check if we observerNode received the right messages from the right peers.
	// Just two messages were send so we iterate 2 times
	var m string
	for i := 0; i < 2; i++ {
		msg := <-msgCh

		err := json.Unmarshal(msg.Data, &m)
		s.Require().NoError(err)
		if msg.GetFrom() == s.nodeAB.host.ID() {
			zlog.Sugar().Debug(msg.GetFrom(), m)
			s.Equal(m, s.nodeAB.msg)
		} else if msg.GetFrom() == s.nodeXY.host.ID() {
			zlog.Sugar().Debug(msg.GetFrom(), m)
			s.Equal(m, s.nodeXY.msg)
		}
	}

	zlog.Sugar().Debug("Finalizing")
	s.ctx.Deadline()
}

// TestPubSubClose tests if peers are successfully quitting and unsubscribing
// a given topic
func (s *PubSubTestSuite) TestPubSubClose() {
	s.NoError(s.observerNode.topicSub.Close(context.Background()))
	s.Equal(0, len(s.observerNode.pubsub.GetTopics()))

	s.NoError(s.nodeAB.topicSub.Close(context.Background()))
	s.Equal(0, len(s.nodeAB.pubsub.GetTopics()))

	s.NoError(s.nodeXY.topicSub.Close(context.Background()))
	s.Equal(0, len(s.nodeXY.pubsub.GetTopics()))
}

func TestPubSubSuite(t *testing.T) {
	suite.Run(t, new(PubSubTestSuite))
}

// startupHostPubSubTest initializes a libp2p host, a GossipSub router and connects
// to given peers
func startupHostPubSubTest(ctx context.Context, peers ...host.Host) (*psNodeConfigTest, error) {
	var err error

	host, err := libp2p.New()
	if err != nil {
		return nil, err
	}

	gs, err := libp2pPS.NewGossipSub(ctx, host)
	if err != nil {
		return nil, err
	}

	if err := connectToPeers(host, peers...); err != nil {
		return nil, err
	}

	return &psNodeConfigTest{
		host: host,
		pubsub: &PubSub{
			gs,
			host.ID().String(),
			sync.Once{}},
	}, nil
}

func connectToPeers(h host.Host, peers ...host.Host) error {
	for _, p := range peers {
		pAddrs := host.InfoFromHost(p)
		err := h.Connect(context.Background(), *pAddrs)
		if err != nil {
			zlog.Sugar().Debug("Couldn't connect to peer")
			return err
		}
		time.Sleep(1 * time.Second)
	}
	zlog.Sugar().Debugf("Peers connected: %v", h.Network().Peers())
	return nil
}
