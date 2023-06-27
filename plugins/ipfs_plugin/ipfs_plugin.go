package ipfs_plugin

import (
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
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

func (p *IPFSPlugin) Start(errCh chan error) {
	err := PullImage(ipfsPluginImg)
	if err != nil {
		zlog.Sugar().Errorf("Couldn't pull ipfs-plugin docker image: %v", err)
		errCh <- err
		return
	}

	zlog.Info("Entering IPFS-Plugin container")

	containerConfig, hostConfig, err := configureContainer(ipfsPluginImg, portDc, addrDc, port)
	if err != nil {
		zlog.Sugar().Errorf("Error occured when configuring container: %v", err)
		errCh <- err
		return
	}

	resp, err := dc.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		zlog.Sugar().Errorf("Unable to create plugin container: %v", err)
		errCh <- err
		return
	}

	zlog.Info("Starting IPFS-Plugin container")
	if err := dc.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		zlog.Sugar().Errorf("Unable to start plugin container: %v", err)
		errCh <- err
		return
	}

	// statusCh, errCh := dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	// TODO: update DHT of usage of resources
	errCh <- nil
	return
}

func configureContainer(img string, exposedPort, hostIP, hostPort string) (*container.Config, *container.HostConfig, error) {
	port, err := nat.NewPort("tcp", exposedPort)
	if err != nil {
		return nil, nil, err
	}

	containerConfig := &container.Config{
		Image: img,
		ExposedPorts: nat.PortSet{
			port: struct{}{},
		},
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
func PullImage(imageName string) error {
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