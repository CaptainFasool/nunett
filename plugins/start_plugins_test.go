package plugins

import (
	"reflect"
	"testing"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/plugins/ipfs_plugin"
)

func TestGetPluginType(t *testing.T) {
	tests := []struct {
		testName        string
		pluginName      string
		expectedErr     bool
		pluginInterface reflect.Type
	}{
		{
			testName:        "ipfs-plugin",
			pluginName:      "ipfs-plugin",
			expectedErr:     false,
			pluginInterface: reflect.TypeOf(&ipfs_plugin.IPFSPlugin{}),
		},
		{
			testName:        "unknown-plugin",
			pluginName:      "unknown-plugin",
			expectedErr:     true,
			pluginInterface: nil,
		},
		{
			testName:        "empty string",
			pluginName:      "",
			expectedErr:     true,
			pluginInterface: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
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

func TestGetMetadataPlugins(t *testing.T) {
	expectedPlugins := []string{"ipfs-plugin", "test-plugin"}

	mockReadMetadataFile := func() (models.MetadataV2, error) {
		return models.MetadataV2{
			Plugins: expectedPlugins,
		}, nil
	}

	plugins, err := getMetadataPlugins(mockReadMetadataFile)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(plugins) != len(expectedPlugins) {
		t.Errorf("Wanted %d plugins, received %d plugins", len(expectedPlugins), len(plugins))
	}

	for i, plugin := range plugins {
		if expectedPlugins[i] != plugin {
			t.Errorf("Wanted plugin: %s, received: %s", expectedPlugins[i], plugin)
		}
	}
}
