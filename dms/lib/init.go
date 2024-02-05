package library

import (
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"gitlab.com/nunet/device-management-service/telemetry/klogger"
	"gitlab.com/nunet/device-management-service/telemetry/logger"
)

var zlog otelzap.Logger

func init() {
	zlog = logger.OtelZapLogger("library")
	klogger.InitializeLogger(klogger.LogLevelWarning)

}
