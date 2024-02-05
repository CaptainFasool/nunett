package messaging

import (
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"gitlab.com/nunet/device-management-service/telemetry/logger"
)

var (
	zlog                 otelzap.Logger
	ProgressFilePathOnCP = make(chan string, 1)
)

func init() {
	zlog = logger.OtelZapLogger("internal/messaging")
}
