package plugins_management

import (
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/telemetry"
)

type PluginsInfoChannels struct {
	ResourcesCh chan models.Resources
	ErrCh       chan error
}

type Plugin interface {
	Run(*PluginsInfoChannels)
	Stop(*PluginsInfoChannels) error
	IsRunning(*PluginsInfoChannels) (bool, error)
	OnboardedName() string
}

// ManagePlugins manages all the plugins which include DHT updates for resources
// usage for every one of them that started successfully
func ManagePlugins(pluginsCentralChannels *PluginsInfoChannels) {
	for {
		select {
		case resourcesUsage := <-pluginsCentralChannels.ResourcesCh:
			zlog.Sugar().Debug("Updating FreeResources as startup of plugin")

			hardwareResources, err := telemetry.NewHardwareResources()
			if err != nil {
				zlog.Sugar().Error(err)
			}

			hardwareResources.IncreaseFreeResources(resourcesUsage)
			hardwareResources.UpdateDBFreeResources()
		case err := <-pluginsCentralChannels.ErrCh:
			zlog.Sugar().Error(err)
		}
	}
}
