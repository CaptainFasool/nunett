package oracle

import "gitlab.com/nunet/device-management-service/telemetry/logger"

var zlog *logger.Logger

func init() {
	zlog = logger.New("oracle")
}
