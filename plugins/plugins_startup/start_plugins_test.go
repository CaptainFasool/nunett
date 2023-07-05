package plugins_startup

import (
	"reflect"
	"testing"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/plugins/ipfs_plugin"
	"gitlab.com/nunet/device-management-service/utils"
)

type expectedMockPlugin struct {
	pluginName      string
	expectedErr     bool
	pluginInterface reflect.Type
}

func getMockExpectedPluginsTable() map[string]expectedMockPlugin {
	return map[string]expectedMockPlugin{
		"ipfs-plugin": {
			pluginName:      "ipfs-plugin",
			expectedErr:     false,
			pluginInterface: reflect.TypeOf(&ipfs_plugin.IPFSPlugin{}),
		},
		"unknown-plugin": {
			pluginName:      "unknown-plugin",
			expectedErr:     true,
			pluginInterface: nil,
		},
		"empty-string": {
			pluginName:      "",
			expectedErr:     true,
			pluginInterface: nil,
		},
	}
}

func getMockExpectedPluginsNames() []string {
	pluginsTable := getMockExpectedPluginsTable()
	var pluginsNames []string
	for _, pluginTestInfo := range pluginsTable {
		if len(pluginTestInfo.pluginName) == 0 {
			continue
		}
		pluginsNames = append(pluginsNames, pluginTestInfo.pluginName)
	}
	return pluginsNames
}

func retrieveEnabledPluginsFromMock() []reflect.Type {
	pluginsTable := getMockExpectedPluginsTable()
	var plugins []reflect.Type
	for _, pluginInfo := range pluginsTable {
		if pluginInfo.pluginInterface != nil {
			plugins = append(plugins, pluginInfo.pluginInterface)
		}
	}
	return plugins
}

func TestGetPluginType(t *testing.T) {
	pluginsTable := getMockExpectedPluginsTable()
	for testName, tt := range pluginsTable {
		t.Run(testName, func(t *testing.T) {
			pluginInterface, err := getPluginType(tt.pluginName)
			if reflect.TypeOf(pluginInterface) != tt.pluginInterface {
				t.Errorf("getPluginType() = %v, want %v", reflect.TypeOf(pluginInterface), tt.pluginInterface)
			}

			if (err != nil) != tt.expectedErr {
				t.Errorf("getPluginType() error = %v, expectedErr %v", err, tt.expectedErr)
			}
			return
		})
	}
}

func mockReadMetadataFile() (models.MetadataV2, error) {
	expectedPlugins := getMockExpectedPluginsNames()
	return models.MetadataV2{
		Plugins: expectedPlugins,
	}, nil
}

func TestGetMetadataPlugins(t *testing.T) {
	expectedPlugins := getMockExpectedPluginsNames()
	plugins, err := getMetadataPlugins(mockReadMetadataFile)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(plugins) != len(expectedPlugins) {
		t.Errorf("Wanted %d plugins, received %d plugins", len(expectedPlugins), len(plugins))
	}

	if !utils.AreSlicesEqual(plugins, expectedPlugins) {
		t.Errorf("Wanted plugins: %s, received plugins: %s", expectedPlugins, plugins)
	}
}

func TestSolveEnabledPlugins(t *testing.T) {
	expectedEnabledPlugins := retrieveEnabledPluginsFromMock()

	plugins, err := solveEnabledPlugins(mockReadMetadataFile)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(plugins) != len(expectedEnabledPlugins) {
		t.Errorf("Wanted %d plugins, received %d plugins", len(expectedEnabledPlugins), len(plugins))
	}

	// Transforming plugins to reflect.Type to be compared with expected plugins
	var pluginsByType []reflect.Type
	for _, pluginInterface := range plugins {
		pluginsByType = append(pluginsByType, reflect.TypeOf(pluginInterface))
	}

	// Checking if we got the expected plugins types
	// TODO: maybe improve this part
	matchedPlugins := 0
OuterLoop:
	for _, expectedPlugin := range expectedEnabledPlugins {
		for _, enabledPlugin := range plugins {
			if reflect.TypeOf(enabledPlugin) == expectedPlugin {
				matchedPlugins++
				continue OuterLoop
			}
		}
	}

	if matchedPlugins != len(expectedEnabledPlugins) {
		t.Errorf("got %v, want %v", plugins, expectedEnabledPlugins)
	}
}
