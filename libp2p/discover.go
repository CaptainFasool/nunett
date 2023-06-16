package libp2p

import (
	"context"
	"os"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

func (p2p DMSp2p) StartDiscovery(ctx context.Context, rendezvous string) {
	Discover(ctx, p2p.Host, p2p.DHT, rendezvous)
}

func (p2p P2P) StartDiscovery(ctx context.Context, rendezvous string) {
	Discover(ctx, p2p.Host, p2p.DHT, rendezvous)
}

func Discover(ctx context.Context, node host.Host, idht *dht.IpfsDHT, rendezvous string) {

	var routingDiscovery = drouting.NewRoutingDiscovery(idht)
	dutil.Advertise(ctx, routingDiscovery, rendezvous)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			peers, err := dutil.FindPeers(ctx, routingDiscovery, rendezvous)
			if err != nil {
				zlog.Sugar().Errorf("failed to discover peers: %v", err)
			}
			peers = filterAddrs(peers)
			for _, p := range peers {
				if p.ID == node.ID() {
					continue
				}
				if node.Network().Connectedness(p.ID) != network.Connected {
					_, err = node.Network().DialPeer(ctx, p.ID)
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
	}
}

func (p2p DMSp2p) getPeers(ctx context.Context, rendezvous string) ([]peer.AddrInfo, error) {

	routingDiscovery := drouting.NewRoutingDiscovery(p2p.DHT)
	dutil.Advertise(ctx, routingDiscovery, rendezvous)
	peers, err := dutil.FindPeers(ctx, routingDiscovery, rendezvous)
	if err != nil {
		zlog.Sugar().Errorf("Error Finding Peers: %s\n", err.Error())
	}
	peers = filterAddrs(peers)

	return peers, nil
}

func (p2p P2P) getPeers(ctx context.Context, rendezvous string) ([]peer.AddrInfo, error) {

	routingDiscovery := drouting.NewRoutingDiscovery(p2p.DHT)
	dutil.Advertise(ctx, routingDiscovery, rendezvous)
	peers, err := dutil.FindPeers(ctx, routingDiscovery, rendezvous)
	if err != nil {
		zlog.Sugar().Errorf("Error Finding Peers: %s\n", err.Error())
	}
	peers = filterAddrs(peers)

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
