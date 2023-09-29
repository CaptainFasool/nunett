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
			p2p.peers = discoverPeers(ctx, p2p.Host, p2p.DHT, rendezvous)
			p2p.dialPeers(ctx)
		}
	}
}

func (p2p DMSp2p) dialPeers(ctx context.Context) {
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
}

func discoverPeers(ctx context.Context, node host.Host, idht *dht.IpfsDHT, rendezvous string) []peer.AddrInfo {
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
		zlog.Sugar().Errorf("failed to discover peers: %v", err)
	}
	peers = filterAddrs(peers)
	zlog.Sugar().Debugf("Discover - found peers: %v", peers)
	return peers
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

func dialPeersContinuously(ctx context.Context, h host.Host, idht *dht.IpfsDHT, peers []peer.ID) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, id := range peers {
				if err := findAndDialPeer(ctx, h, idht, id); err != nil {
					zlog.Sugar().Debugf(
						"Unexpected error when finding or dialing the peer %v, Error: %v",
						id, err)
				}

			}
		}
	}
}

func findAndDialPeer(ctx context.Context, h host.Host, idht *dht.IpfsDHT, id peer.ID) error {
	if h.Network().Connectedness(id) != network.Connected {
		addrs, err := idht.FindPeer(ctx, id)
		if err != nil {
			return fmt.Errorf("Couldn't find peer, Error: %w", err)
		}
		_, err = h.Network().DialPeer(ctx, addrs.ID)
		if err != nil {
			return fmt.Errorf("Couldn't dial peer, Error: %w", err)
		}
	}
	return nil
}
