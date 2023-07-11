package machines

import (
	"github.com/libp2p/go-libp2p/core/host"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
)

// FilterPeers searches for available compute providers given specific parameters in depReq.
func FilterPeers(depReq models.DeploymentRequest, node host.Host) []models.PeerData {
	machines := libp2p.FetchMachines(node)
	zlog.Sugar().Debugf("DHT machines: %v", machines)

	var peers []models.PeerData

	for _, val := range machines {
		peers = append(peers, val)
	}

	peers = libp2p.PeersWithMatchingSpec(peers, depReq)
	if depReq.ServiceType == "ml-training-gpu" {
		if depReq.Params.MachineType == "gpu" {
			peers = libp2p.PeersWithGPU(peers)
			return peers
		}
	}

	if depReq.ServiceType == "cardano_node" {
		peers = libp2p.PeersWithCardanoAllowed(peers)
	}

	return peers
}
