package ipfs_plugin

import (
	"context"

	"github.com/docker/docker/client"
	"gitlab.com/nunet/device-management-service/internal/logger"
)

var (
	ctx  context.Context
	dc   *client.Client
	zlog *logger.Logger
)

func init() {
	var err error

	zlog = logger.New("ipfs_plugin")
	ctx = context.Background()

	dc, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		zlog.Sugar().Errorf("Unable to initialize Docker client for ipfs-plugin: %v", err)
		return
	}
}
