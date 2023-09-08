package pubsub

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p"
	libp2pPS "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/stretchr/testify/suite"
	"gitlab.com/nunet/device-management-service/utils"
)

type PubSubTestSuite struct {
	suite.Suite
	ctx          context.Context
	node1        *psNodeConfigTest
	observerNode *psNodeConfigTest
	topicName    string
	wg           *sync.WaitGroup
}

type psNodeConfigTest struct {
	host   host.Host
	pubsub *PubSubPeer
	topic  *PsTopicSubscription
	msg    string
}

func startupHostPubSubTest(ctx context.Context, peers ...host.Host) (*psNodeConfigTest, error) {
	var err error

	zlog.Sugar().Debug("Creating host")

	port, err := utils.GetFreePort()
	if err != nil {
		return nil, err
	}

	host, err := libp2p.New(
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port),
		),
	)

	if err != nil {
		return nil, err
	}

	ps, err := newTestGossipPubSub(ctx, host)
	if err != nil {
		return nil, err
	}

	if err := connectToPeers(host, peers...); err != nil {
		return nil, err
	}

	return &psNodeConfigTest{
		host:   host,
		pubsub: ps,
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

func (s *PubSubTestSuite) SetupSuite() {
	var err error
	s.ctx = context.Background()

	s.observerNode, err = startupHostPubSubTest(context.Background())
	s.Require().NoError(err)
	time.Sleep(5 * time.Second)

	s.node1, err = startupHostPubSubTest(context.Background(), s.observerNode.host)
	s.Require().NoError(err)

	s.topicName = "topicTest"
	s.node1.msg = "I love lasagna"

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
}

// TestPubSub tests the libp2p's GossipSub implementation and message publishing
// and retrieving processes
func (s *PubSubTestSuite) TestPubSub() {
	var err error

	s.observerNode.topic, err = s.observerNode.pubsub.JoinTopic(s.topicName)
	s.Require().NoError(err)
	s.observerNode.topic.topic.EventHandler()
	go s.observerNode.topic.listenForMessages(s.ctx)
	time.Sleep(1 * time.Second)

	s.node1.topic, err = s.node1.pubsub.JoinTopic(s.topicName)
	s.Require().NoError(err)
	time.Sleep(1 * time.Second)

	zlog.Sugar().Debugf("topics: %v", s.observerNode.pubsub)
	zlog.Sugar().Debugf("topics: %v", s.node1.pubsub)
	zlog.Sugar().Debug(s.observerNode.topic.topic.ListPeers(), s.node1.topic.topic.ListPeers())

	//msgCh := make(chan *libp2pPS.Message)
	//go s.observerNode.topic.listenForMessages(context.Background(), msgCh)
	time.Sleep(1 * time.Second)

	s.NoError(s.observerNode.topic.Publish("dsadsa"))
	s.NoError(s.node1.topic.Publish(s.node1.msg))
	time.Sleep(1 * time.Second)

	// Check if we observerNode received the right messages from the right peers
	// Just to messages were send so we iterate 2 times
	zlog.Sugar().Debug(s.observerNode.topic.topic.ListPeers(), s.node1.topic.topic.ListPeers())

	zlog.Sugar().Debug("Finalizing")

	s.ctx.Deadline()
}

func TestPubSubSuite(t *testing.T) {
	suite.Run(t, new(PubSubTestSuite))
}
