package libp2p

import (
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/internal/logger"
	"gitlab.com/nunet/device-management-service/models"
)

const (
	// Stream Protocols
	// Stream Protocol for Deployment Requests
	DepReqProtocolID = "/nunet/dms/depreq/0.0.2"

	// Stream Protocol for Chat
	ChatProtocolID = "/nunet/dms/chat/0.0.1"

	// Stream Protocol for Ping
	PingProtocolID = "/nunet/dms/ping/0.0.1"

	// namespaces
	// Custom namespace for DHT protocol
	customNamespace = "/nunet-dht/"

	// Rendezvous Points
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

var (
	zlog             otelzap.Logger
	kadPrefix        = dht.ProtocolPrefix("/nunet")
	gettingDHTUpdate = false

	// bootstrap peers provided by NuNet
	NuNetBootstrapPeers []multiaddr.Multiaddr
)

var (
	DepReqQueue       = make(chan models.DeploymentRequest)
	DepResQueue       = make(chan models.DeploymentResponse)
	newPeer           = make(chan peer.AddrInfo)
	newPeers          = make(chan []peer.AddrInfo)
	resultChan        = make(chan models.PeerData)
	JobLogStderrQueue = make(chan string)
	JobLogStdoutQueue = make(chan string)
	JobFailedQueue    = make(chan string)
	JobCompletedQueue = make(chan string)
)

func init() {
	zlog = logger.OtelZapLogger("libp2p")

	for _, s := range config.GetConfig().P2P.BootstrapPeers {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			panic(err)
		}
		NuNetBootstrapPeers = append(NuNetBootstrapPeers, ma)
	}
}
