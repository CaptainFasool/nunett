package onboarding

import (
	"testing"
)

func TestOnboardPlugins(t *testing.T) {
	onboardedPluginsNames := []string{"Plugin1", "Plugin2"}

	plugins, err := onboardPlugins(onboardedPluginsNames)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(plugins) != 2 {
		t.Fatalf("Expected 2 plugins, got %d", len(plugins))
	}

	plugin1 := plugins[0]
	if plugin1.Name != "ipfs-plugin" {
		t.Errorf("Expected name to be 'ipfs-plugin', got '%s'", plugin1.Name)
	}

	// Add more checks for the other fields and the other plugin as needed
}

func aTestOnboardPlugins(t *testing.T) {
	// Test case 1: Successful onboarding of plugins
	onboardedPluginsNames := []string{"plugin1", "plugin2"}
	expectedOnboardedPlugins := []models.Plugin{
		{Name: "plugin1", Version: "1.0"},
		{Name: "plugin2", Version: "2.0"},
	}

	onboardedPlugins, err := onboardPlugins(onboardedPluginsNames)
	assert.NoError(t, err)
	assert.Equal(t, expectedOnboardedPlugins, onboardedPlugins)

	// Test case 2: No plugins onboarded
	onboardedPluginsNames = []string{}
	expectedOnboardedPlugins = []models.Plugin{}

	onboardedPlugins, err = onboardPlugins(onboardedPluginsNames)
	assert.NoError(t, err)
	assert.Equal(t, expectedOnboardedPlugins, onboardedPlugins)

	// Test case 3: Error decoding plugins.toml file
	onboardedPluginsNames = []string{"plugin1", "plugin2"}
	expectedOnboardedPlugins = nil

	onboardedPlugins, err = onboardPlugins(onboardedPluginsNames)
	assert.Error(t, err)
	assert.Equal(t, expectedOnboardedPlugins, onboardedPlugins)
}
