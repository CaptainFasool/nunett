package libp2p

import (
	"context"
	"errors"
	"time"

	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2pdiscovery "github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	bt "gitlab.com/nunet/device-management-service/internal/background_tasks"
	"gitlab.com/nunet/device-management-service/models"
)

// Libp2p contains the configuration for a Libp2p instance.
type Libp2p struct {
	Host      host.Host
	DHT       *dht.IpfsDHT
	PS        peerstore.Peerstore
	peers     []peer.AddrInfo
	discovery libp2pdiscovery.Discovery

	// register services
	pingService *ping.PingService

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
		config: config,
	}, nil
}

// Init initializes a libp2p host with its dependencies.
func (l *Libp2p) Init(context context.Context) error {
	host, dht, err := NewHost(context, l.config)
	if err != nil {
		zlog.Sugar().Error(err)
		return err
	}

	l.Host = host
	l.DHT = dht
	l.PS = host.Peerstore()
	l.discovery = drouting.NewRoutingDiscovery(dht)

	return nil
}

// Start performs network bootstrapping, peer discovery and protocols handling.
func (l *Libp2p) Start(context context.Context) error {
	// advertise randevouz discovery
	err := l.advertiseForRendezvousDiscovery(context)
	if err != nil {
		zlog.Sugar().Errorf("failed to start network with randevouz discovery: %v", err)
	}

	// bootstrap should return error if it had an error
	err = l.Bootstrap(context, l.config.BootstrapPeers)
	if err != nil {
		zlog.Sugar().Errorf("failed to start network: %v", err)
		return err
	}

	// discover
	err = l.DiscoverDialPeers(context)
	if err != nil {
		zlog.Sugar().Errorf("failed to discover peers: %v", err)
	}

	// set stream handlers
	l.registerStreamHandlers()

	// register period peer discoveryTask task
	discoveryTask := &bt.Task{
		Name:        "Peer Discovery",
		Description: "Periodic task to discover new peers every 15 minutes",
		Function: func(args interface{}) error {
			return l.DiscoverDialPeers(context)
		},
		Triggers: []bt.Trigger{&bt.PeriodicTrigger{Interval: 15 * time.Minute}},
	}
	l.config.Scheduler.AddTask(discoveryTask)

	// handlers
	return nil
}

func (l *Libp2p) registerStreamHandlers() {
	l.pingService = ping.NewPingService(l.Host)
	l.Host.SetStreamHandler(protocol.ID("/ipfs/ping/1.0.0"), l.pingService.PingHandler)
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
	return l.Host.Close()
}

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

func (l *Libp2p) Advertise(adId string, data []byte) error {
	return nil
}

func (l *Libp2p) Unadvertise(adId string) error {
	return nil
}

func (l *Libp2p) Publish(topic string, data []byte) error {
	return nil
}

func (l *Libp2p) Subscribe(topic string, handler func(data []byte)) error {
	return nil
}
