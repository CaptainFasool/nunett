package plugins_management

import (
	"errors"

	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/firecracker/telemetry"
	"gitlab.com/nunet/device-management-service/models"
	"gorm.io/gorm"
)

type PluginsInfoChannels struct {
	SucceedStartup chan *models.PluginInfo
	ErrCh          chan error
}

type Plugin interface {
	Run(*PluginsInfoChannels)
	Stop(*PluginsInfoChannels) error
	IsRunning(*PluginsInfoChannels) (bool, error)
}

// ManagePlugins manages all the plugins which include DHT updates for resources
// usage for every one of them that started successfully
func ManagePlugins(pluginsCentralChannels *PluginsInfoChannels) {
	for {
		select {
		case plugin := <-pluginsCentralChannels.SucceedStartup:
			zlog.Sugar().Infof("Plugin %v successfully started", plugin.Name)
			zlog.Sugar().Debugf("Updating DB record for plugin %v", plugin.Name)
			err := upsertPluginInfo(plugin)
			if err != nil {
				continue
			}
			zlog.Sugar().Debugf("Updating free resources with resource usage of plugin %v", plugin.Name)
			err = telemetry.CalcFreeResources()
			if err != nil {
				zlog.Sugar().Errorf("Couldn't update free resources: %v", err)
				continue
			}

		case err := <-pluginsCentralChannels.ErrCh:
			zlog.Sugar().Error(err)
		}
	}
}

func upsertPluginInfo(plugin *models.PluginInfo) error {
	pluginModel := models.PluginInfo{}
	res := db.DB.Where(models.PluginInfo{Name: plugin.Name}).Take(&pluginModel)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			// Record for plugin not found
			resCreation := db.DB.Create(plugin)
			if resCreation.Error != nil {
				zlog.Sugar().Errorf("Error when creating a record for Plugins: %w", res.Error)
				return resCreation.Error
			}
		} else {
			// Error querying plugin table
			zlog.Sugar().Errorf("Error when querying Plugins table: %w", res.Error)
			return res.Error
		}
	} else {
		// Plugin record found, update it with new info
		resUpdating := db.DB.Model(&pluginModel).Updates(plugin)
		if resUpdating.Error != nil {
			zlog.Sugar().Errorf("Error when updating a record for Plugins: %w", res.Error)
			return resUpdating.Error
		}
	}

	return nil
}
