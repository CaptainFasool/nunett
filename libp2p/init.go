package libp2p

import (
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
