package ipfs_plugin

import (
	"github.com/docker/docker/api/types"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/plugins/plugins_management"
)

type IPFSPlugin struct {
	info models.PluginInfo
}

const (
	ipfsPluginImg = "registry.gitlab.com/nunet/data-persistence/ipfs-plugin:0.0.1"
	pluginName    = "ipfs-plugin"
	addrDc        = "127.0.0.1"
	portDc        = "31001"
)

func NewIPFSPlugin() *IPFSPlugin {
	p := &IPFSPlugin{}
	i := models.PluginInfo{}
	i.Name = pluginName
	i.ResourcesUsage.TotCpuHz = 1000
	i.ResourcesUsage.Ram = 4000

	p.info = i
	return p
}

// Run implements the plugin interface and deals with the startup of IPFS-Plugin,
// downloading the image, configuring and starting the Docker container
func (p *IPFSPlugin) Run(pluginsManager *plugins_management.PluginsInfoChannels) {
	err := plugins_management.PullImage(ipfsPluginImg, dc)
	if err != nil {
		zlog.Sugar().Errorf("Couldn't pull ipfs-plugin docker image: %v", err)
		pluginsManager.ErrCh <- err
		return
	}

	zlog.Sugar().Debug("Entering IPFS-Plugin container")

	containerConfig, hostConfig, err := plugins_management.ConfigureContainer(
		ipfsPluginImg, portDc, addrDc, port, p.info.Name, dc)
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

	pluginsManager.SucceedStartup <- &p.info

	// statusCh, errCh := dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
}

// Stop stops the IPFS-Plugin Docker container and return an error if any.
func (p *IPFSPlugin) Stop(pluginsManager *plugins_management.PluginsInfoChannels) error {
	err := plugins_management.StopPluginDcContainer(p.info.Name, dc)
	if err != nil {
		pluginsManager.ErrCh <- err
		return err
	}
	return nil
}

// IsRunning checks if a IPFS-Plugin Docker container is running
func (p *IPFSPlugin) IsRunning(pluginsManager *plugins_management.PluginsInfoChannels) (bool, error) {
	isRunning, err := plugins_management.IsPluginDcContainerRunning(p.info.Name, dc)
	if err != nil {
		pluginsManager.ErrCh <- err
		return false, err
	}

	if isRunning {
		return true, nil
	}
	return false, nil
}
