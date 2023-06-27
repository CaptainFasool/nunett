package machines

import (
	"github.com/libp2p/go-libp2p/core/host"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
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

	peers = filterByNeededPlugins(peers, depReq)

	zlog.Sugar().Infof("Number of peers with matched requirements: %v", len(peers))

	return peers
}

// filterByNeededPlugins filters the peers by the necessary plugins that a compute provider
// must be running based on the deployment request params
func filterByNeededPlugins(peers []models.PeerData, depReq models.DeploymentRequest) []models.PeerData {
	// TODO: Some plugins will run along with DMS when starting DMS,
	// other plugins will start accordingly to the initiation of jobs.
	// This implementation only consider plugins which run along DMS.
	var neededPlugins []string
	var peersWithNeededPlugins []models.PeerData

	// Check of needed plugins (Improve this to be dinamically for several plugins)
	if isIPFSPLuginNeeded(depReq) {
		neededPlugins = append(neededPlugins, "ipfs-plugin")
	}

	if len(neededPlugins) == 0 {
		return peers
	}

	for _, peer := range peers {
		if utils.SliceContainsSlice(neededPlugins, peer.EnabledPlugins) {
			peersWithNeededPlugins = append(peersWithNeededPlugins, peer)
		}
	}

	zlog.Sugar().Infof("Needed plugins that compute provider must be running: %v", neededPlugins)
	return peersWithNeededPlugins
}

// isIPFSPLuginNeeded returns peers with IPFS-Plugin enabled
func isIPFSPLuginNeeded(depReq models.DeploymentRequest) bool {
	pluginIPFSFunctionalities := [...]string{"outputIPFS"}
	for _, functionality := range pluginIPFSFunctionalities {
		if utils.SliceContainsValue(functionality, depReq.Params.AdditionalFeatures) {
			return true
		}
	}
	return false
}
