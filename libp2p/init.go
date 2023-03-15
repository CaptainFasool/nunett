package libp2p

import (
	"github.com/multiformats/go-multiaddr"
	"gitlab.com/nunet/device-management-service/internal/logger"
	"gitlab.com/nunet/device-management-service/models"
)

var zlog *logger.Logger

func init() {
	zlog = logger.New("libp2p")
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
		"/ip4/159.69.44.89/tcp/6763/p2p/QmQ2irHa8aFTLRhkbkQCRrounE4MbttNp8ki7Nmys4F9NP",
		"/ip4/5.78.68.130/tcp/6763/p2p/Qmf16N2ecJVWufa29XKLNyiBxKWqVPNZXjbL3JisPcGqTw",
		"/ip4/5.161.196.44/tcp/6763/p2p/QmTkWP72uECwCsiiYDpCFeTrVeUM9huGTPsg3m6bHxYQFZ",
	} {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			panic(err)
		}
		NuNetBootstrapPeers = append(NuNetBootstrapPeers, ma)
	}
}
