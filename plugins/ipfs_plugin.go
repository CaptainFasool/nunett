package plugins

import (
	dockerDMS "gitlab.com/nunet/device-management-service/docker"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

var (
	ipfsPluginImg = "registry.gitlab.com/nunet/data-persistence/ipfs-plugin:0.0.1"
)

func RunContainer() {
	err := dockerDMS.PullImage(ipfsPluginImg)
	if err != nil {
		zlog.Sugar().Errorf("Couldn't pull ipfs-plugin image: %v", err)
		return
	}

	zlog.Info("Entering plugin container")

	containerConfig := &container.Config{
		Image: ipfsPluginImg,
	}

	resp, err := dc.ContainerCreate(ctx, containerConfig, nil, nil, nil, "")

	if err != nil {
		zlog.Sugar().Errorf("Unable to create plugin container: %v", err)
		return
	}

	if err := dc.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		zlog.Sugar().Errorf("Unable to start plugin container: %v", err)
		return
	}

	// statusCh, errCh := dc.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

}
