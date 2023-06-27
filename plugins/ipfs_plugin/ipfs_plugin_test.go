package ipfs_plugin

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"reflect"
	"testing"
)

func TestConfigureContainer(t *testing.T) {
	testCases := []struct {
		name           string
		img            string
		exposedPort    string
		hostIP         string
		hostPort       string
		expectedConfig *container.Config
		expectedHost   *container.HostConfig
		expectErr      bool
	}{
		{
			name:        "Test with valid input",
			img:         "test-image",
			exposedPort: "8080",
			hostIP:      "127.0.0.1",
			hostPort:    "8080",
			expectedConfig: &container.Config{
				Image: "test-image",
				ExposedPorts: nat.PortSet{
					"8080/tcp": struct{}{},
				},
			},
			expectedHost: &container.HostConfig{
				PortBindings: nat.PortMap{
					"8080/tcp": []nat.PortBinding{
						{
							HostIP:   "127.0.0.1",
							HostPort: "8080",
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name:        "Test with invalid port",
			img:         "test-image",
			exposedPort: "invalid",
			hostIP:      "127.0.0.1",
			hostPort:    "8080",
			expectErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, host, err := configureContainer(tc.img, tc.exposedPort, tc.hostIP, tc.hostPort)
			if (err != nil) != tc.expectErr {
				t.Errorf("Expected error: %v, got: %v", tc.expectErr, err)
			}
			if !reflect.DeepEqual(config, tc.expectedConfig) {
				t.Errorf("Expected config: %v, got: %v", tc.expectedConfig, config)
			}
			if !reflect.DeepEqual(host, tc.expectedHost) {
				t.Errorf("Expected host config: %v, got: %v", tc.expectedHost, host)
			}
		})
	}
}
