package ipfs_plugin

import (
	"fmt"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/plugins/plugins_management"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type IPFSPlugin struct {
	info        models.PluginInfo
	exposedPort string
	exposedAddr string
	dockerImg   string
}

const (
	allInterfacesIP = "0.0.0.0"
	loopbackIP      = "127.0.0.1"

	kuboDockerImg   = "ipfs/kubo:latest"
	kuboSwarmPort   = "4001"
	kuboAPIPort     = "5001"
	kuboGatewayPort = "8080"
)

func NewIPFSPlugin() *IPFSPlugin {
	p := &IPFSPlugin{}
	p.dockerImg = "registry.gitlab.com/nunet/data-persistence/ipfs-plugin:0.0.1"
	p.exposedAddr = loopbackIP
	p.exposedPort = "31001"

	i := models.PluginInfo{}
	i.Name = "ipfs-plugin"
	i.ResourcesUsage.TotCpuHz = 1000
	i.ResourcesUsage.Ram = 4000

	p.info = i
	return p
}

// Run deals with the startup of IPFS-Plugin, downloading the image,
// configuring and starting the Docker container
func (p *IPFSPlugin) Run(pluginsManager *plugins_management.PluginsInfoChannels) {
	err := pullImage(p.dockerImg, dc)
	if err != nil {
		zlog.Sugar().Errorf("Couldn't pull ipfs-plugin docker image: %v", err)
		pluginsManager.ErrCh <- err
		return
	}

	zlog.Sugar().Debug("Entering IPFS-Plugin container")

	containerConfig, hostConfig, err := p.configureContainer()
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

	// We're not running from within IPFS-Plugin container because this
	// requires some workarounds with Docker socket and the container itself
	zlog.Sugar().Debug("Starting Kubo node container")
	if err := runKuboNode(dc); err != nil {
		zlog.Sugar().Errorf("Unable to start Kubo node for IPFS-Plugin: %v", err)
		pluginsManager.ErrCh <- err
		return
	}

	pluginsManager.SucceedStartup <- &p.info

	// statusCh, errCh := dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
}

// Stop stops the IPFS-Plugin Docker container and return an error if any.
func (p *IPFSPlugin) Stop(pluginsManager *plugins_management.PluginsInfoChannels) error {
	err := stopPluginDcContainer(p.info.Name, dc)
	if err != nil {
		pluginsManager.ErrCh <- err
		return err
	}
	return nil
}

// IsRunning checks if a IPFS-Plugin Docker container is running
func (p *IPFSPlugin) IsRunning(pluginsManager *plugins_management.PluginsInfoChannels) (bool, error) {
	isRunning, err := isPluginDcContainerRunning(p.info.Name, dc)
	if err != nil {
		pluginsManager.ErrCh <- err
		return false, err
	}

	if isRunning {
		return true, nil
	}
	return false, nil
}

// configureContainer returns the configuration values for IPFS-Plugin Docker container.
// Here is where ports, imageURL and other Docker container configs are defined.
func (p *IPFSPlugin) configureContainer() (*container.Config, *container.HostConfig, error) {
	env := []string{
		fmt.Sprintf("kuboSwarmPort=%v", kuboSwarmPort),
		fmt.Sprintf("kuboAPIPort=%v", kuboAPIPort),
		fmt.Sprintf("kuboGatewayPort=%v", kuboGatewayPort),
	}

	port, err := nat.NewPort("tcp", p.exposedPort)
	if err != nil {
		return nil, nil, err
	}

	labels := map[string]string{
		"dms-related": "true",
	}

	labels["dms-plugin"] = p.info.Name

	containerConfig := &container.Config{
		Image: p.dockerImg,
		ExposedPorts: nat.PortSet{
			port: struct{}{},
		},
		Labels: labels,
		Env:    env,
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			port: []nat.PortBinding{
				{
					HostIP:   p.exposedAddr,
					HostPort: p.exposedPort,
				},
			},
		},
	}
	return containerConfig, hostConfig, nil
}

// runKuboNode starts Kubo node (docker container) through its official docker image
func runKuboNode(dc *client.Client) error {
	// Temporary Note:
	// The following is the same as the command:
	// docker run -d --name ipfs_host -e IPFS_PROFILE=server
	// -p 4001:4001 -p 4001:4001/udp -p 127.0.0.1:8080:8080 -p 127.0.0.1:5001:5001 ipfs/kubo:latest
	containerName := "ipfs_host"

	if err := stopAndRemoveContainerIfExists(containerName, dc); err != nil {
		return fmt.Errorf("Couldn't stop and remove container, Error: %w", err)
	}

	err := pullImage(kuboDockerImg, dc)
	if err != nil {
		return fmt.Errorf("Couldn't pull Kubo docker image: %v", err)
	}

	// Configure container environment variables
	env := []string{"IPFS_PROFILE=server"}

	natSwarmPortTCP := fmt.Sprintf("%v/tcp", kuboSwarmPort)
	natSwarmPortUDP := fmt.Sprintf("%v/udp", kuboSwarmPort)
	natGatewayPort := fmt.Sprintf("%v/tcp", kuboGatewayPort)
	natAPIPort := fmt.Sprintf("%v/tcp", kuboAPIPort)

	// Configure container ports
	portBindings := nat.PortMap{
		nat.Port(natSwarmPortTCP): []nat.PortBinding{
			{HostIP: allInterfacesIP, HostPort: kuboSwarmPort},
		},
		nat.Port(natSwarmPortUDP): []nat.PortBinding{
			{HostIP: allInterfacesIP, HostPort: kuboSwarmPort},
		},
		nat.Port(natGatewayPort): []nat.PortBinding{
			{HostIP: loopbackIP, HostPort: kuboGatewayPort},
		},
		nat.Port(natAPIPort): []nat.PortBinding{
			{HostIP: loopbackIP, HostPort: kuboAPIPort},
		},
	}

	config := &container.Config{
		Image: kuboDockerImg,
		Env:   env,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
	}

	resp, err := dc.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		return fmt.Errorf("Unable to create Kubo container: %v", err)
	}

	err = dc.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("Unable to start Kubo container: %v", err)
	}

	// statusCh, errCh := dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	return nil
}
