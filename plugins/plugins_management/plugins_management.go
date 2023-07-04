package plugins_management

import (
	"fmt"

	"gitlab.com/nunet/device-management-service/models"
)

type PluginsInfoChannels struct {
	ResourcesCh chan models.FreeResources
	ErrCh       chan error
}

// pluginsManager manages all the plugins which include DHT updates for resources
// usage for every one of them that started successfully
func ManagePlugins(pluginsCentralChannels *PluginsInfoChannels) {
	for {
		select {
		case resourcesUsage := <-pluginsCentralChannels.ResourcesCh:
			fmt.Print(resourcesUsage.ID)
		case err := <-pluginsCentralChannels.ErrCh:
			fmt.Print(err)
		}
	}
}
