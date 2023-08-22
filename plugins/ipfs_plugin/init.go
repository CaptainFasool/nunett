package ipfs_plugin

import (
	"context"

	"gitlab.com/nunet/device-management-service/internal/logger"
)

var (
	ctx  context.Context
	zlog *logger.Logger
)

func init() {
	zlog = logger.New("ipfs_plugin")
	ctx = context.Background()
}
