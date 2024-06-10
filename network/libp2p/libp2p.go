package libp2p

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ipfs/go-cid"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
	"google.golang.org/protobuf/proto"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pdiscovery "github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	bt "gitlab.com/nunet/device-management-service/internal/background_tasks"
	"gitlab.com/nunet/device-management-service/models"
	commonproto "gitlab.com/nunet/device-management-service/proto/generated/v1/common"
)

// Libp2p contains the configuration for a Libp2p instance.
type Libp2p struct {
	Host         host.Host
	DHT          *dht.IpfsDHT
	PS           peerstore.Peerstore
	pubsub       *pubsub.PubSub
	pubsubTopics map[string]*pubsub.Topic

	// a list of peers discovered by discovery
	discoveredPeers []peer.AddrInfo
	discovery       libp2pdiscovery.Discovery

	// services
	pingService *ping.PingService

	// tasks
	discoveryTask *bt.Task

	config *models.Libp2pConfig
}

// New creates a libp2p instance.
func New(config *models.Libp2pConfig) (*Libp2p, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	if config.Scheduler == nil {
		return nil, errors.New("scheduler is nil")
	}

	return &Libp2p{
		config:       config,
		pubsubTopics: make(map[string]*pubsub.Topic),
	}, nil
}

// Init initializes a libp2p host with its dependencies.
func (l *Libp2p) Init(context context.Context) error {
	host, dht, pubsub, err := NewHost(context, l.config)
	if err != nil {
		zlog.Sugar().Error(err)
		return err
	}

	l.Host = host
	l.DHT = dht
	l.PS = host.Peerstore()
	l.discovery = drouting.NewRoutingDiscovery(dht)
	l.pubsub = pubsub

	return nil
}

// Start performs network bootstrapping, peer discovery and protocols handling.
func (l *Libp2p) Start(context context.Context) error {
	// set stream handlers
	l.registerStreamHandlers()

	// bootstrap should return error if it had an error
	err := l.Bootstrap(context, l.config.BootstrapPeers)
	if err != nil {
		zlog.Sugar().Errorf("failed to start network: %v", err)
		return err
	}

	// advertise randevouz discovery
	err = l.advertiseForRendezvousDiscovery(context)
	if err != nil {
		zlog.Sugar().Errorf("failed to start network with randevouz discovery: %v", err)
	}

	// discover
	err = l.DiscoverDialPeers(context)
	if err != nil {
		zlog.Sugar().Errorf("failed to discover peers: %v", err)
	}

	// register period peer discoveryTask task
	discoveryTask := &bt.Task{
		Name:        "Peer Discovery",
		Description: "Periodic task to discover new peers every 15 minutes",
		Function: func(args interface{}) error {
			return l.DiscoverDialPeers(context)
		},
		Triggers: []bt.Trigger{&bt.PeriodicTrigger{Interval: 15 * time.Minute}},
	}

	l.discoveryTask = l.config.Scheduler.AddTask(discoveryTask)

	return nil
}

// SendMessage sends a message to a list of peers.
func (l *Libp2p) SendMessage(ctx context.Context, addrs []string, msg []byte) error {
	return errors.New("unimplemented")
}

// GetMultiaddr returns the peer's multiaddr.
func (l *Libp2p) GetMultiaddr() ([]multiaddr.Multiaddr, error) {
	peerInfo := peer.AddrInfo{
		ID:    l.Host.ID(),
		Addrs: l.Host.Addrs(),
	}
	return peer.AddrInfoToP2pAddrs(&peerInfo)
}

// Stop performs a cleanup of any resources used in this package.
func (l *Libp2p) Stop() error {
	var errorMessages []string

	l.config.Scheduler.RemoveTask(l.discoveryTask.ID)

	if err := l.DHT.Close(); err != nil {
		errorMessages = append(errorMessages, err.Error())
	}
	if err := l.Host.Close(); err != nil {
		errorMessages = append(errorMessages, err.Error())
	}

	if len(errorMessages) > 0 {
		return errors.New(strings.Join(errorMessages, "; "))
	}

	return nil
}

// Stat returns the status about the libp2p network.
func (l *Libp2p) Stat() models.NetworkStats {
	return models.NetworkStats{}
}

// Ping the remote address. The remote address is the encoded peer id which will be decoded and used here.
func (l *Libp2p) Ping(ctx context.Context, peerIDAddress string, timeout time.Duration) (models.PingResult, error) {
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	remotePeer, err := peer.Decode(peerIDAddress)
	if err != nil {
		return models.PingResult{}, err
	}

	pingChan := ping.Ping(pingCtx, l.Host, remotePeer)

	select {
	case res := <-pingChan:
		return models.PingResult{
			RTT:     res.RTT,
			Success: true,
		}, nil
	case <-pingCtx.Done():
		return models.PingResult{
			Error: pingCtx.Err(),
		}, pingCtx.Err()
	}
}

// Advertisements return all the advertisements in the network related to a key.
// The network is queried to find providers for the given key, and peers which we aren't connected to
// can be retrieved.
func (l *Libp2p) Advertisements(ctx context.Context, key string) ([]*commonproto.Advertisement, error) {
	if key == "" {
		return nil, errors.New("advertisement key is empty")
	}

	customCID, err := createCIDFromKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cid for key %s: %w", key, err)
	}

	addrInfo, err := l.DHT.FindProviders(ctx, customCID)
	if err != nil {
		return nil, fmt.Errorf("failed to find providers for key %s: %w", key, err)
	}
	var advertisements []*commonproto.Advertisement
	for _, v := range addrInfo {
		// TODO: use go routines to get the values in parallel.
		bytesAdvertisement, err := l.DHT.GetValue(ctx, l.getCustomNamespace(key, v.ID.String()))
		if err != nil {
			continue
		}
		var ad commonproto.Advertisement
		if err := proto.Unmarshal(bytesAdvertisement, &ad); err != nil {
			return nil, fmt.Errorf("failed to unmarshal advertisement payload: %w", err)
		}
		advertisements = append(advertisements, &ad)
	}

	return advertisements, nil
}

// Advertise given data and a key pushes the data to the dht.
func (l *Libp2p) Advertise(ctx context.Context, key string, data []byte) error {
	if key == "" {
		return errors.New("advertisement key is empty")
	}

	pubKeyBytes, err := l.getPublicKey()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	envelope := &commonproto.Advertisement{
		PeerId:    l.Host.ID().String(),
		Timestamp: time.Now().Unix(),
		Data:      data,
		PublicKey: pubKeyBytes,
	}

	concatenatedBytes := bytes.Join([][]byte{
		[]byte(envelope.PeerId),
		{byte(envelope.Timestamp)},
		envelope.Data,
		pubKeyBytes,
	}, nil)

	sig, err := l.sign(concatenatedBytes)
	if err != nil {
		return fmt.Errorf("failed to sign advertisement envelope content: %w", err)
	}

	envelope.Signature = sig

	envelopeBytes, err := proto.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal advertise envelope: %w", err)
	}

	customCID, err := createCIDFromKey(key)
	if err != nil {
		return fmt.Errorf("failed to create cid for key %s: %w", key, err)
	}

	err = l.DHT.PutValue(ctx, l.getCustomNamespace(key, l.DHT.PeerID().String()), envelopeBytes)
	if err != nil {
		return fmt.Errorf("failed to put key %s into the dht: %w", key, err)
	}

	err = l.DHT.Provide(ctx, customCID, true)
	if err != nil {
		return fmt.Errorf("failed to provide key %s into the dht: %w", key, err)
	}

	return nil
}

// Unadvertise removes the data from the dht.
func (l *Libp2p) Unadvertise(ctx context.Context, key string) error {
	err := l.DHT.PutValue(ctx, l.getCustomNamespace(key, l.DHT.PeerID().String()), nil)
	if err != nil {
		return fmt.Errorf("failed to remove key %s from the DHT: %w", key, err)
	}

	return nil
}

// Publish publishes data to a topic.
// The requirements are that only one topic handler should exist per topic.
func (l *Libp2p) Publish(ctx context.Context, topic string, data []byte) error {
	topicHandler, ok := l.pubsubTopics[topic]
	if !ok {
		t, err := l.pubsub.Join(topic)
		if err != nil {
			return fmt.Errorf("failed to join topic %s: %w", topic, err)
		}
		topicHandler = t
		l.pubsubTopics[topic] = topicHandler
	}

	err := topicHandler.Publish(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", topic, err)
	}

	return nil
}

// Subscribe subscribes to a topic and sends the messages to the handler.
func (l *Libp2p) Subscribe(ctx context.Context, topic string, handler func(data []byte)) error {
	topicHandler, ok := l.pubsubTopics[topic]
	if !ok {
		t, err := l.pubsub.Join(topic)
		if err != nil {
			return fmt.Errorf("failed to join topic %s: %w", topic, err)
		}
		topicHandler = t
		l.pubsubTopics[topic] = topicHandler
	}

	sub, err := topicHandler.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				continue
			}
			handler(msg.Data)
		}
	}()

	return nil
}

func (l *Libp2p) registerStreamHandlers() {
	l.pingService = ping.NewPingService(l.Host)
	l.Host.SetStreamHandler(protocol.ID("/ipfs/ping/1.0.0"), l.pingService.PingHandler)
}

func (l *Libp2p) sign(data []byte) ([]byte, error) {
	privKey := l.Host.Peerstore().PrivKey(l.Host.ID())
	if privKey == nil {
		return nil, errors.New("private key not found for the host")
	}

	signature, err := privKey.Sign(data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return signature, nil
}

func (l *Libp2p) getPublicKey() ([]byte, error) {
	privKey := l.Host.Peerstore().PrivKey(l.Host.ID())
	if privKey == nil {
		return nil, errors.New("private key not found for the host")
	}

	pubKey := privKey.GetPublic()
	return pubKey.Raw()
}

func (l *Libp2p) getCustomNamespace(key, peerID string) string {
	return fmt.Sprintf("%s-%s-%s", l.config.CustomNamespace, key, peerID)
}

func createCIDFromKey(key string) (cid.Cid, error) {
	hash := sha256.Sum256([]byte(key))
	mh, err := multihash.Encode(hash[:], multihash.SHA2_256)
	if err != nil {
		return cid.Cid{}, err
	}
	return cid.NewCidV1(cid.Raw, mh), nil
}
