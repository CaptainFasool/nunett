package plugins

import "fmt"

type plugin interface {
	start(chan error) // TODO: pass also as a param the model.peerInfo
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

	enabledPlugins, err := getEnablePlugins()
	if err != nil {
		zlog.Sugar().Errorf("Couldn't get enable plugins: %v", err)
		return
	}


	errCh := make(chan error)
	go pluginsManager(errCh)

	var currentPlugin plugin

	for _, pluginName := range enabledPlugins {
		currentPlugin, err = getPluginType(pluginName)
		if err != nil {
			zlog.Sugar().Errorf(err.Error())
			return
		}
		go currentPlugin.start(errCh)
	}

}

func pluginsManager(errCh chan error) {
	i := <-errCh
	fmt.Println(i)
}

// getEnablePlugins retrieves from the DB the plugins enabled by the user.
func getEnablePlugins() ([]string, error) {
	// TODO (get plugins from user local DB)
	enablePlugins := []string{"ipfs-plugin"}
	return enablePlugins, nil
}

// getPluginType returns, based on the plugin name, the specific plugin type struct
// which can implement different interface methods
func getPluginType(pluginName string) (plugin, error) {
	switch pluginName {
	case "ipfs-plugin":
		return &IPFSPlugin{}, nil
	default:
		return nil, fmt.Errorf("Plugin name wrong or not implemented on DMS side")
	}
}
