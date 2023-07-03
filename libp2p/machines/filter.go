package machines

import (
	"github.com/libp2p/go-libp2p/core/host"
	"gitlab.com/nunet/device-management-service/libp2p"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
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

	peers = filterByNeededPlugins(peers, depReq)

	zlog.Sugar().Infof("Number of peers with matched requirements: %v", len(peers))

	return peers
}

// filterByNeededPlugins filters the peers by the necessary plugins that a compute provider
// must be running based on the deployment requested params and features
func filterByNeededPlugins(peers []models.PeerData, depReq models.DeploymentRequest) []models.PeerData {
	// TODO: Some plugins will run along with DMS when starting DMS,
	// other plugins will start accordingly to the initiation of jobs.
	// This implementation only consider plugins which run along DMS.
	var peersWithNeededPlugins []models.PeerData

	neededPlugins := solvePluginsNeeded(depReq)

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

// solvePluginsNeeded returns related plugins based on requested functionalities
func solvePluginsNeeded(depReq models.DeploymentRequest) []string {
	// Some AdditionalFeatures may not be necessarily plugins,
	// so we might not rely on relating AdditionalFeatures to plugins.
	// This might return 0 peers when in reality the AdditionalFeature was
	// not a plugin at all, just a normal AdditionalFeature
	// TLDR: change this on webapp side probably (UI or parsing within the react code)
	var neededPlugins []string
	pluginsAdded := make(map[string]bool)

	relationFuncsPlugins := getRelationFuncsPlugins()
	for _, functionality := range depReq.Params.AdditionalFeatures {
		plugin, ok := relationFuncsPlugins[functionality]
		if _, added := pluginsAdded[plugin]; !added && ok {
			neededPlugins = append(neededPlugins, plugin)
			pluginsAdded[plugin] = true
		}
	}

	return neededPlugins
}

// getRelationFuncsPlugins returns a map which relating each functionality requested (by the SP) to a
// given plugin. Being a key:value pair, this is the structure <functionality>:<Plugin>
func getRelationFuncsPlugins() map[string]string {
	relation := map[string]string{
		"outputIPFS":      "ipfs-plugin",
		"jobResumingIPFS": "ipfs-plugin",
	}

	return relation
}
