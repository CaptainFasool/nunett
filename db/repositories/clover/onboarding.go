package repositories_clover

import (
	"github.com/ostafen/clover/v2"
	"gitlab.com/nunet/device-management-service/db/repositories"
	"gitlab.com/nunet/device-management-service/models"
)

// LogBinAuthRepositoryClover is a Clover implementation of the LogBinAuthRepository interface.
type LogBinAuthRepositoryClover struct {
	repositories.GenericEntityRepository[models.LogBinAuth]
}

// NewLogBinAuthRepository creates a new instance of LogBinAuthRepositoryClover.
// It initializes and returns a Clover-based repository for LogBinAuth entity.
func NewLogBinAuthRepository(db *clover.DB) repositories.LogBinAuthRepository {
	return &LogBinAuthRepositoryClover{
		NewGenericEntityRepository[models.LogBinAuth](db),
	}
}
