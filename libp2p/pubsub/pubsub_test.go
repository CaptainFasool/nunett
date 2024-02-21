package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
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
	ctxCancel    context.CancelFunc
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
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())

	s.topicName = "test topic"

	s.observerNode, err = startupHostPubSubTest(s.ctx, "")
	s.Require().NoError(err)

	s.nodeAB, err = startupHostPubSubTest(s.ctx, s.topicName, s.observerNode.host)
	s.Require().NoError(err)

	s.nodeXY, err = startupHostPubSubTest(s.ctx, s.topicName, s.observerNode.host)
	s.Require().NoError(err)

	// put 20 more nodes
	for i := 0; i < 20; i++ {
		_, err = startupHostPubSubTest(s.ctx, s.topicName, s.observerNode.host)
		s.Require().NoError(err)
	}

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
	s.observerNode.topicSub, err = s.observerNode.pubsub.JoinSubscribeTopic(s.ctx, s.topicName, true)
	s.Require().NoError(err)
	msgCh := make(chan *libp2pPS.Message)
	go s.observerNode.topicSub.ListenForMessages(context.Background(), msgCh)
	time.Sleep(1 * time.Second)

	// Publishing message with observerNode which should be ignored by
	// its own ListenForMessages()
	s.NoError(s.observerNode.topicSub.Publish(s.observerNode.msg))
	time.Sleep(1 * time.Second)

	time.Sleep(60 * 2 * time.Second)

	fmt.Printf("observer node known nodes: %v\n", s.observerNode.pubsub.ListPeers(s.topicName))
	fmt.Printf("nodeAB known nodes: %v\n", s.nodeAB.pubsub.ListPeers(s.topicName))
	fmt.Printf("nodeXY known nodes: %v\n", s.nodeXY.pubsub.ListPeers(s.topicName))

	// print connected nodes based on host
	fmt.Printf("observer node connected peers: %v\n", s.observerNode.host.Network().Peers())
	fmt.Printf("nodeAB connected peers: %v\n", s.nodeAB.host.Network().Peers())
	fmt.Printf("nodeXY connected peers: %v\n", s.nodeXY.host.Network().Peers())

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
	close(msgCh)
}

// TestPubSubClose tests if peers are successfully quitting and unsubscribing
// a given topic
func (s *PubSubTestSuite) TestPubSubClose() {
	s.NoError(s.nodeAB.topicSub.Close(context.Background()))
	s.Equal(0, len(s.nodeAB.pubsub.GetTopics()))

	s.NoError(s.nodeXY.topicSub.Close(context.Background()))
	s.Equal(0, len(s.nodeXY.pubsub.GetTopics()))

	s.NoError(s.observerNode.topicSub.Close(context.Background()))
	s.Equal(0, len(s.observerNode.pubsub.GetTopics()))

}

func TestPubSubSuite(t *testing.T) {
	suite.Run(t, new(PubSubTestSuite))
}

// startupHostPubSubTest initializes a libp2p host, a GossipSub router and connects
// to given peers
func startupHostPubSubTest(ctx context.Context, psTopic string, peers ...host.Host) (*psNodeConfigTest, error) {
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

	ps := &PubSub{
		gs,
		host.ID().String(),
		sync.Once{},
		map[string]*PsTopicSubscription{},
	}

	if psTopic == "" {
		return &psNodeConfigTest{
			host:   host,
			pubsub: ps,
		}, nil
	}

	topicSub, err := ps.JoinSubscribeTopic(ctx, psTopic, false)
	if err != nil {
		return nil, err
	}

	return &psNodeConfigTest{
		host:     host,
		pubsub:   ps,
		topicSub: topicSub,
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
