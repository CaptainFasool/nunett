package plugins_startup

import (
	"fmt"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/plugins/ipfs_plugin"
	"gitlab.com/nunet/device-management-service/plugins/plugins_management"
	"gitlab.com/nunet/device-management-service/utils"
)

type ReadMetadataFileFunc func() (models.MetadataV2, error)

func StartPlugins() {
	zlog.Sugar().Debug("Starting plugins")

	enabledPlugins, err := solveEnabledPlugins(utils.ReadMetadataFile)
	if err != nil {
		zlog.Sugar().Warn("Enabled plugins were not started: Couldn't get enabled plugins: ", err)
		return
	}

	if enabledPlugins == nil {
		zlog.Sugar().Info("No plugins enabled")
		return
	}

	pluginsCentralChannels := &plugins_management.PluginsInfoChannels{
		SucceedStartup: make(chan *models.PluginInfo),
		ErrCh:          make(chan error),
	}
	go plugins_management.ManagePlugins(pluginsCentralChannels)

	for _, currentPlugin := range enabledPlugins {
		// Currently killing plugin if already running to update possible new configs
		if isRunning, _ := currentPlugin.IsRunning(pluginsCentralChannels); isRunning {
			currentPlugin.Stop(pluginsCentralChannels)
		}
		go currentPlugin.Run(pluginsCentralChannels)
	}

	zlog.Sugar().Debug("Exiting StartPlugins")
	return
}

// solveEnabledPlugins gets enabled plugins within metadata and solve their types
func solveEnabledPlugins(readMetadataFile ReadMetadataFileFunc) ([]plugins_management.Plugin, error) {
	strPlugins, err := getMetadataPlugins(readMetadataFile)
	if err != nil {
		return []plugins_management.Plugin{}, err
	}

	var enabledPlugins []plugins_management.Plugin
	for _, pluginName := range strPlugins {
		pluginType, err := GetPluginType(pluginName)
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
func GetPluginType(pluginName string) (plugins_management.Plugin, error) {
	switch pluginName {
	case "ipfs-plugin":
		return ipfs_plugin.NewIPFSPlugin(), nil
	default:
		return nil, fmt.Errorf("Plugin name wrong or not implemented on DMS side: %v", pluginName)
	}
}
