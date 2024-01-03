package tokenomics

import "gitlab.com/nunet/device-management-service/internal/logger"

var zlog *logger.Logger

const (
	transactionWithdrawnStatus   = "withdrawn"
	transactionRefundedStatus    = "refunded"
	transactionDistributedStatus = "distributed"
)

func init() {
	zlog = logger.New("tokenomics")
}
