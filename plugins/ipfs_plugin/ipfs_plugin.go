package ipfs_plugin

import (
	"fmt"
	"os"
	"os/exec"

	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/plugins/plugins_management"
)

type IPFSPlugin struct {
	info    models.PluginInfo
	process *os.Process
	running bool
	port    string
	addr    string
}

const (
	loopbackIP = "127.0.0.1"
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
	pluginsManager.SucceedStartup <- &p.info
	zlog.Sugar().Infof("Plugin %v started, path: %v, pid: %v", p.info.Name, cmd.Path, p.process.Pid)

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

	err := p.process.Kill()
	if err != nil {
		pluginsManager.ErrCh <- fmt.Errorf("Unable to kill ipfs-plugin process, Erro: %w", err)
		return err
	}

	p.process = nil
	p.running = false
	return nil
}

// IsRunning checks if a IPFS-Plugin Docker container is running
func (p *IPFSPlugin) IsRunning(pluginsManager *plugins_management.PluginsInfoChannels) (bool, error) {
	return p.running, nil
}
