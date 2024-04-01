package libp2p

import (
	"context"
	"fmt"
	"os"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

var err error

func (p2p DMSp2p) StartDiscovery(ctx context.Context, rendezvous string) {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			zlog.Debug("Discovery - context done")
			return
		case <-stopDiscovery:
			zlog.Debug("Discovery - stop")
			ticker.Stop()
			return
		case <-ticker.C:
			p2p.peers, err = discoverPeers(ctx, p2p.Host, p2p.DHT, rendezvous)
			if err != nil {
				zlog.Sugar().Errorln(err)
				return
			}
			p2p.peers = filterAddrs(p2p.peers)
			err = p2p.dialPeers(ctx)
			if err != nil {
				zlog.Sugar().Errorln(err)
				return
			}
		}
	}
}

func (p2p DMSp2p) dialPeers(ctx context.Context) error {
	for _, p := range p2p.peers {
		newPeer <- p
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

func discoverPeers(ctx context.Context, node host.Host, idht *dht.IpfsDHT, rendezvous string) ([]peer.AddrInfo, error) {
	var routingDiscovery = drouting.NewRoutingDiscovery(idht)
	dutil.Advertise(ctx, routingDiscovery, rendezvous)

	zlog.Debug("Discover - searching for peers")
	peers, err := dutil.FindPeers(
		ctx,
		routingDiscovery,
		rendezvous,
		discovery.Limit(40),
	)
	if err != nil {
		return []peer.AddrInfo{}, fmt.Errorf("failed to discover peers: %v", err)
	}
	zlog.Sugar().Debugf("Discover - found peers: %v", peers)
	return peers, nil
}

func filterAddrs(peers []peer.AddrInfo) []peer.AddrInfo {
	var filtered []peer.AddrInfo
	for _, p := range peers {
		if p.ID == p2p.Host.ID() {
			continue
		}
		if len(p.Addrs) == 0 {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered
}
