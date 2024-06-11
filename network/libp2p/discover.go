package libp2p

import (
	"context"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

var err error

func (p *Libp2p) DiscoverDialPeers(ctx context.Context) error {
	p.peers, err = p.findPeers(ctx)
	if err != nil {
		return err
	}

	// filter out peers with no listening addresses and self host
	filterSpec := NoAddrIDFilter{ID: p.Host.ID()}
	p.peers = PeerPassFilter(p.peers, filterSpec)

	err = p.dialPeers(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (p2p Libp2p) dialPeers(ctx context.Context) error {
	for _, p := range p2p.peers {
		if p.ID == p2p.Host.ID() {
			continue
		}
		if p2p.Host.Network().Connectedness(p.ID) != network.Connected {
			_, err := p2p.Host.Network().DialPeer(ctx, p.ID)
			if err != nil {
				if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
					zlog.Sugar().Debugf("couldn't establish connection with: %s - error: %v", p.ID.String(), err)
				}
				continue
			}
			if _, debugMode := os.LookupEnv("NUNET_DEBUG_VERBOSE"); debugMode {
				zlog.Sugar().Debugf("connected with: %s", p.ID.String())
			}

		}
	}
	return nil
}

func (p Libp2p) findPeers(ctx context.Context) ([]peer.AddrInfo, error) {
	var routingDiscovery = drouting.NewRoutingDiscovery(p.DHT)
	dutil.Advertise(ctx, routingDiscovery, p.config.Rendezvous)

	zlog.Debug("Discover - searching for peers")
	peers, err := dutil.FindPeers(
		ctx,
		routingDiscovery,
		p.config.Rendezvous,
		discovery.Limit(40),
	)
	if err != nil {
		return []peer.AddrInfo{}, fmt.Errorf("failed to discover peers: %v", err)
	}
	zlog.Sugar().Debugf("Discover - found peers: %v", peers)
	return peers, nil
}
