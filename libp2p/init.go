package libp2p

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/logger"
	"gitlab.com/nunet/device-management-service/models"
)

var zlog otelzap.Logger

func init() {
	zlog = logger.OtelZapLogger("libp2p")
}

const (
	// Stream Protocol for DHT
	DHTProtocolID = "/nunet/dms/dht/0.0.2"

	// Stream Protocol for Deployment Requests
	DepReqProtocolID = "/nunet/dms/depreq/0.0.2"

	// Stream Protocol for Chat
	ChatProtocolID = "/nunet/dms/chat/0.0.1"

	// Stream Protocol for Ping
	PingProtocolID = "/nunet/dms/ping/0.0.1"
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

var relayPeer = make(chan peer.AddrInfo)

// bootstrap peers provided by NuNet
var NuNetBootstrapPeers []multiaddr.Multiaddr

func init() {
	for _, s := range config.GetConfig().P2P.BootstrapPeers {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			panic(err)
		}
		NuNetBootstrapPeers = append(NuNetBootstrapPeers, ma)
	}
}
