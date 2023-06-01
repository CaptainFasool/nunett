package plugins

import (
	dockerDMS "gitlab.com/nunet/device-management-service/docker"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

type IPFSPlugin struct{}

const (
	ipfsPluginImg = "registry.gitlab.com/nunet/data-persistence/ipfs-plugin:0.0.1"
)

func (p *IPFSPlugin) start(errCh chan error) {
	err := dockerDMS.PullImage(ipfsPluginImg)
	if err != nil {
		zlog.Sugar().Errorf("Couldn't pull ipfs-plugin image: %v", err)
		errCh <- err
		return
	}

	zlog.Info("Entering plugin container")

	containerConfig := &container.Config{
		Image: ipfsPluginImg,
		ExposedPorts: nat.PortSet{
			"31001/tcp": struct{}{},
		},
	}

	// hostConfig := &container.HostConfig{
	// 	PortBindings: nat.PortMap{
	// 		"31001/tcp": []nat.PortBinding{
	// 			{
	// 				HostIP:   "127.0.0.1",
	// 				HostPort: "31001",
	// 			},
	// 		},
	// 	},
	// }

	resp, err := dc.ContainerCreate(ctx, containerConfig, nil, nil, nil, "")
	if err != nil {
		zlog.Sugar().Errorf("Unable to create plugin container: %v", err)
		errCh <- err
		return
	}

	if err := dc.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		zlog.Sugar().Errorf("Unable to start plugin container: %v", err)
		errCh <- err
		return
	}

	// statusCh, errCh := dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	errCh <- nil
	return

}
