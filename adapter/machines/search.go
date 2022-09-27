package machines

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"gitlab.com/nunet/device-management-service/adapter"
)

// SearchDevice searches for available compute providers given specific parameters.
func SearchDevice(c *gin.Context, serviceType string) ([]adapter.Peer, error) {
	jsonBytes, err := adapter.FetchDht()
	if err != nil {
		return nil, err
	}

	var dht adapter.DHT

	err = json.Unmarshal(jsonBytes, &dht)
	if err != nil {
		return nil, err
	}

	var peers []adapter.Peer
	if serviceType == "gpu" {
		peers = adapter.PeersWithGPU(dht.PeerMeta)
	}

	if serviceType == "cardano" {
		peers = adapter.PeersWithCardanoAllowed(dht.PeerMeta)
	}
	peers = adapter.PeersNonBusy(peers)

	return peers, nil
}
