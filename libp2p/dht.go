package libp2p

import (
	"context"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

func Bootstrap(ctx context.Context, h host.Host, idht *dht.IpfsDHT) error {
	err := idht.Bootstrap(ctx)
	if err != nil {
		return err
	}
	for _, addr := range dht.DefaultBootstrapPeers {
		pi, err := peer.AddrInfoFromP2pAddr(addr)
		// We ignore errors as some bootstrap peers may be down
		// and that is fine.
		if err != nil {
			continue
		}
		err = h.Connect(ctx, *pi)
		if err != nil {
			return err
		}
	}
	return nil

}
