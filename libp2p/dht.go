package libp2p

import (
	"context"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

func autoRelay(ctx context.Context, node host.Host, ndht dht.IpfsDHT) {
	for {
		peers, err := ndht.GetClosestPeers(ctx, node.ID().String())
		if err != nil {
			zlog.Sugar().Infof("GetClosestPeers error: %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		for _, p := range peers {
			addrs := node.Peerstore().Addrs(p)
			if len(addrs) == 0 {
				continue
			}
			make(chan peer.AddrInfo) <- peer.AddrInfo{
				ID:    p,
				Addrs: addrs,
			}
		}
	}
}

func Bootstrap(ctx context.Context, h host.Host, idht *dht.IpfsDHT) error {
	if err := idht.Bootstrap(ctx); err != nil {
		return err
	}

	for i, p := range dht.GetDefaultBootstrapPeerAddrInfos() {
		if err := h.Connect(ctx, p); err != nil {
			zlog.Sugar().Errorf("failed to connect to bootstrap node #%v\n", i)
		} else {
			zlog.Sugar().Infof("Connected to Bootstrap Node #%v\n", i)
		}
	}

	zlog.Info("Done Bootstrapping")

	go autoRelay(ctx, h, *idht)

	return nil
}
