package onboarding

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
)

func onboardPlugins(r io.Reader, onboardedPluginsNames []string) ([]models.Plugin, error) {
	var plugins []models.Plugin
	var onboardedPlugins []models.Plugin

	_, err := toml.NewDecoder(r).Decode(&plugins)
	if err != nil {
		return nil, err
	}

	for _, plugin := range plugins {
		if utils.SliceContainsValue(plugin.Name, onboardedPluginsNames) {
			onboardedPlugins = append(onboardedPlugins, plugin)
		}
	}

	return onboardedPlugins, nil
}

func tomlToReader(pathTOML string) (*os.File, error) {
	file, err := os.Open(pathTOML)
	if err != nil {
		return nil, err
	}
	return file, nil
}
