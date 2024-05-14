package repositories

import (
	"gitlab.com/nunet/device-management-service/models"
)

// LogBinAuthRepository represents a repository for CRUD operations on LogBinAuth entity.
type LogBinAuthRepository interface {
	GenericEntityRepository[models.LogBinAuth]
}
