package ipfs_plugin

import (
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/plugins/plugins_management"
)

type IPFSPlugin struct{}

const (
	ipfsPluginImg = "registry.gitlab.com/nunet/data-persistence/ipfs-plugin:0.0.1"
	pluginName    = "ipfs-plugin"
	addrDc        = "127.0.0.1"
	portDc        = "31001"
)

func (p *IPFSPlugin) OnboardedName() string {
	return pluginName
}

// Start implements the plugin interface and deals with the startup of IPFS-Plugin,
// downloading the image, configuring and starting the container
func (p *IPFSPlugin) Start(pluginsManager *plugins_management.PluginsInfoChannels) {
	err := pullImage(ipfsPluginImg)
	if err != nil {
		zlog.Sugar().Errorf("Couldn't pull ipfs-plugin docker image: %v", err)
		pluginsManager.ErrCh <- err
		return
	}

	zlog.Sugar().Debug("Entering IPFS-Plugin container")

	containerConfig, hostConfig, err := configureContainer(ipfsPluginImg, portDc, addrDc, port, pluginName)
	if err != nil {
		zlog.Sugar().Errorf("Error occured when configuring container: %v", err)
		pluginsManager.ErrCh <- err
		return
	}

	resp, err := dc.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		zlog.Sugar().Errorf("Unable to create plugin container: %v", err)
		pluginsManager.ErrCh <- err
		return
	}

	zlog.Sugar().Debug("Starting IPFS-Plugin container")
	if err := dc.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		zlog.Sugar().Errorf("Unable to start plugin container: %v", err)
		pluginsManager.ErrCh <- err
		return
	}

	// statusCh, errCh := dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
}

// TODO: The following functions should be moved to docker package whenever we move the deployment
// of jobs to service package

// configureContainer return configuration structs to start a Docker container. Here is where
// ports, imageURL and other Docker container configs are defined.
func configureContainer(img, exposedPort, hostIP, hostPort, plugin string) (*container.Config, *container.HostConfig, error) {
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
func pullImage(imageName string) error {
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
