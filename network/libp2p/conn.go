package libp2p

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

// Ping pings the given peer and returns the result along with the context cancel function
func (p Libp2p) Ping(ctx context.Context, targetPeer peer.ID) (<-chan ping.Result, func()) {
	pingCtx, pingCancel := context.WithCancel(ctx)
	pingResult := ping.Ping(pingCtx, p.Host, targetPeer)
	return pingResult, pingCancel
}
