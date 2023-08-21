package ipfs_plugin

import (
	"fmt"
	"os"
	"os/exec"

	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/plugins/plugins_management"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type IPFSPlugin struct {
	info    models.PluginInfo
	process *os.Process
	running bool
	port    string
	addr    string
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
	p.addr = loopbackIP
	p.port = "31001"

	i := models.PluginInfo{}
	i.Name = "ipfs-plugin"
	i.ResourcesUsage.TotCpuHz = 1000
	i.ResourcesUsage.Ram = 4000

	p.info = i
	return p
}

// Run deals with the startup of IPFS-Plugin through exec.Command()
// in which the default path for plugins is $dms-root/plugins/executables
func (p *IPFSPlugin) Run(pluginsManager *plugins_management.PluginsInfoChannels) {
	zlog.Sugar().Debug("Starting ", p.info.Name)
	executablePath := fmt.Sprintf("%v/%v", config.GetConfig().General.PluginsPath, p.info.Name)
	cmd := exec.Command(executablePath)
	env := []string{
		fmt.Sprintf("kuboSwarmPort=%v", kuboSwarmPort),
		fmt.Sprintf("kuboAPIPort=%v", kuboAPIPort),
		fmt.Sprintf("kuboGatewayPort=%v", kuboGatewayPort),
	}
	cmd.Env = append([]string{}, env...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		pluginsManager.ErrCh <- fmt.Errorf(
			"Couldn't execute cmd.Start() for: %v, Error: %w", p.info.Name, err,
		)
		return
	}

	p.running = true
	p.process = cmd.Process
	zlog.Sugar().Infof("Plugin %v started, path: %v, pid: %v", p.info.Name, cmd.Path, p.process.Pid)

	zlog.Sugar().Debug("Starting Kubo node within a Docker container")
	if err := runKuboNode(dc); err != nil {
		pluginsManager.ErrCh <- fmt.Errorf(
			"Unable to start Kubo node for IPFS-Plugin: %w", err,
		)
		return
	}

	pluginsManager.SucceedStartup <- &p.info

	err = cmd.Wait()
	if err != nil {
		p.running = false
		pluginsManager.ErrCh <- fmt.Errorf(
			"Plugin %v exited, Error: %w", p.info.Name, err,
		)
		return
	}

	return

}

// Stop stops the IPFS-Plugin Docker container and return an error if any.
func (p *IPFSPlugin) Stop(pluginsManager *plugins_management.PluginsInfoChannels) error {
	if p.process == nil {
		return fmt.Errorf("There is no assigned process for plugin %v", p.info.Name)
	}

	defer func() {
		p.process = nil
	}()

	err := p.process.Kill()
	if err != nil {
		pluginsManager.ErrCh <- fmt.Errorf("Unable to kill ipfs-plugin process, Erro: %w", err)
		return err
	}
	return nil
}

// IsRunning checks if a IPFS-Plugin Docker container is running
func (p *IPFSPlugin) IsRunning(pluginsManager *plugins_management.PluginsInfoChannels) (bool, error) {
	return p.running, nil
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

	return nil
}
