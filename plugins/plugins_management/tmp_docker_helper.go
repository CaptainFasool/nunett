package plugins_management

import (
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// TODO: The following functions should be moved to docker package whenever we move the deployment
// of jobs to service package

// StopPluginContainer stops a plugin container based on the plugin name.
func StopPluginDcContainer(pluginName string, dc *client.Client) error {
	// Get the list of running containers
	containers, err := dc.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.Arg("label", "dms-plugin="+pluginName)),
	})
	if err != nil {
		zlog.Sugar().Errorf("Unable to get list of containers to stop plugin: %v", pluginName)
		return err
	}

	for _, container := range containers {
		if err := dc.ContainerStop(ctx, container.ID, nil); err != nil {
			zlog.Sugar().Errorf("Unable to stop plugin container: %v", pluginName)
			return err
		}
	}
	zlog.Sugar().Debugf("Stopping container: %v", pluginName)

	return nil
}

// IsPluginDcContainerRunning checks if a Plugin Docker container is running based on its name
func IsPluginDcContainerRunning(pluginName string, dc *client.Client) (bool, error) {
	// Get the list of running containers
	containers, err := dc.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.Arg("label", "dms-plugin="+pluginName)),
	})
	if err != nil {
		zlog.Sugar().Errorf("Unable to check if plugin is running: %v", pluginName)
		return false, err
	}

	zlog.Sugar().Debugf("Checking if container %v is still running", pluginName)
	if len(containers) > 0 {
		zlog.Sugar().Debugf("Plugin Docker container %v is still running", pluginName)
		return true, nil
	}
	return false, nil
}

// ConfigureContainer return configuration structs to start a Docker container. Here is where
// ports, imageURL and other Docker container configs are defined.
func ConfigureContainer(img, exposedPort, hostIP, hostPort, plugin string, dc *client.Client) (*container.Config, *container.HostConfig, error) {
	port, err := nat.NewPort("tcp", exposedPort)
	if err != nil {
		return nil, nil, err
	}

	labels := map[string]string{
		"dms-related": "true",
	}
	if plugin != "" {
		labels["dms-plugin"] = plugin
	}

	containerConfig := &container.Config{
		Image: img,
		ExposedPorts: nat.PortSet{
			port: struct{}{},
		},
		Labels: labels,
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			port: []nat.PortBinding{
				{
					HostIP:   hostIP,
					HostPort: hostPort,
				},
			},
		},
	}
	return containerConfig, hostConfig, nil
}

// PullImage is a wrapper around Docker SDK's function with same name.
// This function is copied from the docker package.
func PullImage(imageName string, dc *client.Client) error {
	// TODO: We should rename the docker package to deployment package
	// and create a new docker package.
	// OR put this image in utils/utils.go
	out, err := dc.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	defer out.Close()
	io.Copy(os.Stdout, out)
	return nil
}
