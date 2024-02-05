package resources

import (
	"gitlab.com/nunet/device-management-service/telemetry/logger"
)

var zlog *logger.Logger

func init() {
	zlog = logger.New("resources")
}
