package machines

import (
	"gitlab.com/nunet/device-management-service/adapter"
	"gitlab.com/nunet/device-management-service/models"
)

// FilterPeers searches for available compute providers given specific parameters in depReq.
func FilterPeers(depReq models.DeploymentRequest) ([]adapter.Peer, error) {
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

	peers = adapter.PeersWithMatchingSpec(peers, depReq)
	if depReq.ServiceType == "ml-training-gpu" {
		peers = adapter.PeersWithGPU(peers)
	}

	if depReq.ServiceType == "cardano_node" {
		peers = adapter.PeersWithCardanoAllowed(peers)
	}

	return peers, nil
}
