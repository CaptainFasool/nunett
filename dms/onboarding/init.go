package onboarding

import (
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"gitlab.com/nunet/device-management-service/internal/klogger"
	"gitlab.com/nunet/device-management-service/internal/logger"
)

var zlog otelzap.Logger

func init() {
	zlog = logger.OtelZapLogger("onboarding")
	klogger.InitializeLogger(klogger.LogLevelWarning)

}