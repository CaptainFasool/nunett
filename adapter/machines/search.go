package machines

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/adapter"
)

// SearchDevice searches for available compute providers given specific parameters.
func SearchDevice(c *gin.Context, serviceType string) ([]adapter.Peer, error) {
	machines, err := adapter.FetchMachines()
	if err != nil {
		return nil, err
	}

	peers := make([]adapter.Peer, len(machines))

	i := 0
	for _, value := range machines {
		peers[i] = value
		i++
	}

	if serviceType == "gpu" {
		peers = adapter.PeersWithGPU(peers)
	}

	if serviceType == "cardano" {
		peers = adapter.PeersWithCardanoAllowed(peers)
	}

	peers = adapter.PeersNonBusy(peers)

	return peers, nil
}
