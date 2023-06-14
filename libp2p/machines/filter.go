package machines

import (
	"github.com/libp2p/go-libp2p/core/host"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
)

// FilterPeers searches for available compute providers given specific parameters in depReq.
func FilterPeers(depReq models.DeploymentRequest, node host.Host) []models.PeerData {
	var machines models.Machines
	machines, err := libp2p.FetchKadMachines()
	zlog.Sugar().Debugf("KAD machines: %v", machines)
	if err != nil {
		zlog.Sugar().Errorf("failed to fetch machines: %w", err)
	}
	machines2 := libp2p.FetchMachines(node)
	zlog.Sugar().Debugf("Peer store machines: %v", machines2)
	for _, val := range machines2 {
		if _, ok := machines[val.PeerID]; !ok {
			machines[val.PeerID] = val
		}
	}

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
