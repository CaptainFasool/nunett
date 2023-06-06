package plugins

import (
	"context"

	"gitlab.com/nunet/device-management-service/internal/logger"
)

var (
	ctx  context.Context
	zlog *logger.Logger
)

func init() {
	zlog = logger.New("plugins")
	ctx = context.Background()
}
