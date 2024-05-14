package repositories

import (
	"gitlab.com/nunet/device-management-service/models"
)

// DeploymentRequestFlatRepository represents a repository for CRUD operations on DeploymentRequestFlat entities.
type DeploymentRequestFlatRepository interface {
	GenericRepository[models.DeploymentRequestFlat]
}
