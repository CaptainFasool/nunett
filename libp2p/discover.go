package libp2p

import (
	"context"
	"fmt"
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
				zlog.Sugar().Fatalf("Error Discovering Peers: %s\n", err.Error())
			}
			for _, p := range peers {
				if p.ID == node.ID() {
					continue
				}
				if node.Network().Connectedness(p.ID) != network.Connected {
					_, err = node.Network().DialPeer(ctx, p.ID)
					if err != nil {
						// fmt.Println("++++========= connectedness with ", p.ID.String(), " failed - ", err.Error())
						continue
					}
					fmt.Println("++++========= connectedness with ", p.ID.String())

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
	return peers, nil
}