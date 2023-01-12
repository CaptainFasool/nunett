package adapter

import (
	"gitlab.com/nunet/device-management-service/internal/logger"
	"gitlab.com/nunet/device-management-service/models"
)

var (
	zlog *logger.Logger

	DepReqQueue = make(chan models.AdapterMessage)
	DepResQueue = make(chan models.AdapterMessage)
)

func init() {
	zlog = logger.New("adapter")
}
