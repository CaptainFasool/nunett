package statsdb

import (
	"gitlab.com/nunet/device-management-service/internal/logger"
)

var zlog *logger.Logger

func init() {
	zlog = logger.New("statsdb/")
}
