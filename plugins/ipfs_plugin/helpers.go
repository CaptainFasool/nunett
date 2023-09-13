package ipfs_plugin

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// TODO: The following functions should be moved to docker package whenever we move the deployment
// of jobs to service package

// getContainerIDIfExists returns the Docker container ID based on its name
func getContainerIDIfExists(name string, dc *client.Client) (string, error) {
	containers, err := dc.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: filters.NewArgs(filters.Arg("name", name)),
		All:     true,
	})
	if err != nil {
		return "", fmt.Errorf(
			"Unable to get list of containers: %v, Error: %w",
			name, err)
	}

	if len(containers) == 0 {
		return "", fmt.Errorf("No container found with the name: %v", name)
	}

	return containers[0].ID, nil
}

// stopAndRemoveContainerIfExists stops and removes a Docker container if
// exists.
func stopAndRemoveContainerIfExists(name string, dc *client.Client) error {
	id, err := getContainerIDIfExists(name, dc)
	if err != nil {
		return fmt.Errorf("Error checking if container is running: %v", err)
	}

	if id != "" {
		container, err := dc.ContainerInspect(context.Background(), id)
		if err != nil {
			return fmt.Errorf("Unable to inspect container with ID: %v, Error: %w", id, err)
		}

		if container.State.Running {
			if err := dc.ContainerStop(context.Background(), id, nil); err != nil {
				return fmt.Errorf(
					"Unable to stop container with ID: %v, Error: %w",
					id, err)
			}
		}

		if err := dc.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{}); err != nil {
			return fmt.Errorf("Unable to remove container: %v, Error: %w", name, err)
		}
	}
	return nil
}

// stopPluginContainer stops a plugin container based on the plugin name.
func stopPluginDcContainer(pluginName string, dc *client.Client) error {
	// Get the list of running containers
	containers, err := dc.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: filters.NewArgs(filters.Arg("label", "dms-plugin="+pluginName)),
	})
	if err != nil {
		zlog.Sugar().Errorf("Unable to get list of containers to stop plugin: %v", pluginName)
		return err
	}

	for _, container := range containers {
		if err := dc.ContainerStop(context.Background(), container.ID, nil); err != nil {
			zlog.Sugar().Errorf("Unable to stop plugin container: %v", pluginName)
			return err
		}
	}
	zlog.Sugar().Debugf("Stopping container: %v", pluginName)

	return nil
}

// isPluginDcContainerRunning checks if a Plugin Docker container is running based on its name
func isPluginDcContainerRunning(pluginName string, dc *client.Client) (bool, error) {
	// Get the list of running containers
	containers, err := dc.ContainerList(context.Background(), types.ContainerListOptions{
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

// pullImage is a wrapper around Docker SDK's function with same name.
// This function is copied from the docker package.
func pullImage(imageName string, dc *client.Client) error {
	// TODO: We should rename the docker package to deployment package
	// and create a new docker package.
	// OR put this image in utils/utils.go
	out, err := dc.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	defer out.Close()
	io.Copy(os.Stdout, out)
	return nil
}
