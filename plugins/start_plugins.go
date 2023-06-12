package plugins

import (
	"fmt"

	"gitlab.com/nunet/device-management-service/plugins/ipfs_plugin"
	"gitlab.com/nunet/device-management-service/utils"
)

type plugin interface {
	Start(chan error) // TODO: pass also as a param the model.peerInfo
	OnboardedName() string
}

// TODOs:

// 1. Check plugins enabled by CP
// 2. Init plugins enabled

// 1. Pull Container Image
// 2. Run Container Image
// 3. Calculate resources usage by plugin
// 4. Update DB with those resources and updated DHT with decreased free/available resources
// 4. (Optional) Do things while container is running
// 5. When job is finished, remove stored IPFS data for the specific job (send /delete call)
// 6. Free resources (delete container image when stopping DMS)

// StartPlugins initiate all plugins enabled by user, creating a new go routine for each plugin.
func StartPlugins() {
	zlog.Info("Starting plugins")

	enabledPlugins, err := solveEnabledPlugins()
	if err != nil {
		zlog.Sugar().Errorf("Couldn't get enabled plugins: ", err)
	}

	if enabledPlugins == nil {
		zlog.Sugar().Info("No plugins enabled")
		return
	}

	errCh := make(chan error)
	go pluginsManager(errCh)

	for _, currentPlugin := range enabledPlugins {
		go currentPlugin.Start(errCh)
	}

	zlog.Info("Exiting StartPlugins")
	return
}

// solveEnabledPlugins gets enabled plugins within metadata and solve their types
func solveEnabledPlugins() ([]plugin, error) {
	strPlugins, err := getMetadataPlugins()
	if err != nil {
		return []plugin{}, err
	}

	var enabledPlugins []plugin
	for _, pluginName := range strPlugins {
		pluginType, err := getPluginType(pluginName)
		if err != nil {
			zlog.Sugar().Errorf(err.Error())
			continue
		}
		zlog.Sugar().Info("Plugin Enabled: ", pluginName)
		enabledPlugins = append(enabledPlugins, pluginType)
	}
	return enabledPlugins, nil
}

// getMetadataPlugins retrieves from metadataV2.json the plugins enabled by the user.
func getMetadataPlugins() ([]string, error) {
	metadata, err := utils.ReadMetadataFile()
	if err != nil {
		zlog.Sugar().Errorf("Couldn't read from metadata file (you probably hadn't onboarded your machine yet): %v", err)
		return []string{}, err
	}
	enabledPlugins := metadata.Plugins
	return enabledPlugins, nil
}

// getPluginType returns, based on the plugin name, the specific plugin type struct
// which can implement different interface methods
func getPluginType(pluginName string) (plugin, error) {
	switch pluginName {
	case "ipfs-plugin":
		return &ipfs_plugin.IPFSPlugin{}, nil
	default:
		return nil, fmt.Errorf("Plugin name wrong or not implemented on DMS side: %v", pluginName)
	}
}
