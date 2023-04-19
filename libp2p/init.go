package libp2p

import (
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/multiformats/go-multiaddr"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"gitlab.com/nunet/device-management-service/internal/logger"
	"gitlab.com/nunet/device-management-service/models"
)

var zlog otelzap.Logger

var sub event.Subscription

func init() {
	zlog = logger.OtelZapLogger("libp2p")
}

const (
	// Stream Protocol for DHT
	DHTProtocolID = "/nunet/dms/dht/0.0.1"

	// Stream Protocol for Deployment Requests
	DepReqProtocolID = "/nunet/dms/depreq/0.0.1"

	// Stream Protocol for Chat
	ChatProtocolID = "/nunet/dms/chat/0.0.1"
)

const (
	// Team Rendezvous
	TeamRendezvous = "nunet-team"

	// Edge Rendezvous
	EdgeRendezvous = "nunet-edge"

	// Test Rendezvous
	TestRendezvous = "nunet-test"

	// Staging Rendezvous
	StagingRendezvous = "nunet-staging"

	// Prod Rendezvous
	ProdRendezvous = "nunet"
)

var DepReqQueue = make(chan models.DeploymentRequest)
var DepResQueue = make(chan models.DeploymentResponse)

// bootstrap peers provided by NuNet
var NuNetBootstrapPeers []multiaddr.Multiaddr

func init() {
	for _, s := range []string{
        "/ip4/5.161.142.32/tcp/6763/p2p/QmZLHBqTYbu9PKmhnGDkFGgMZaF8B9eZG9YSSCVpkdr7kK",
	} {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			panic(err)
		}
		NuNetBootstrapPeers = append(NuNetBootstrapPeers, ma)
	}
}
