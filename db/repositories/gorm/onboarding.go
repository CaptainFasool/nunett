package repositories_gorm

import (
	"gorm.io/gorm"

	"gitlab.com/nunet/device-management-service/db/repositories"
	"gitlab.com/nunet/device-management-service/models"
)

// LogBinAuthRepositoryGORM is a GORM implementation of the LogBinAuthRepository interface.
type LogBinAuthRepositoryGORM struct {
	repositories.GenericEntityRepository[models.LogBinAuth]
}

// NewLogBinAuthRepository creates a new instance of LogBinAuthRepositoryGORM.
// It initializes and returns a GORM-based repository for LogBinAuth entity.
func NewLogBinAuthRepository(db *gorm.DB) repositories.LogBinAuthRepository {
	return &LogBinAuthRepositoryGORM{
		NewGenericEntityRepository[models.LogBinAuth](db),
	}
}
