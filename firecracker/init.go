package firecracker

import (
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"gitlab.com/nunet/device-management-service/internal/logger"
)

var zlog otelzap.Logger

func init() {
	// ctx := context.Background()
	zlog = logger.OtelZapLogger("firecracker")
}
