package plugins

import (
	"fmt"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/plugins/ipfs_plugin"
	"gitlab.com/nunet/device-management-service/utils"
)

type plugin interface {
	Start(chan error) // TODO: pass also as a param the model.peerInfo
	OnboardedName() string
}

type ReadMetadataFileFunc func() (models.MetadataV2, error)

func StartPlugins() {
	zlog.Info("Starting plugins")

	enabledPlugins, err := solveEnabledPlugins(utils.ReadMetadataFile)
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
func solveEnabledPlugins(readMetadataFile ReadMetadataFileFunc) ([]plugin, error) {
	strPlugins, err := getMetadataPlugins(readMetadataFile)
	if err != nil {
		return []plugin{}, err
	}

	var enabledPlugins []plugin
	for _, pluginName := range strPlugins {
		pluginType, err := getPluginType(pluginName)
		if err != nil {
			zlog.Sugar().Warn(err.Error())
			continue
		}
		zlog.Sugar().Info("Plugins Enabled: ", pluginName)
		enabledPlugins = append(enabledPlugins, pluginType)
	}
	return enabledPlugins, nil
}

// getMetadataPlugins retrieves from metadataV2.json the plugins enabled by the user.
func getMetadataPlugins(readMetadataFile ReadMetadataFileFunc) ([]string, error) {
	metadata, err := readMetadataFile()
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
