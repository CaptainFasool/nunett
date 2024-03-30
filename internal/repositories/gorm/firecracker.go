package repositories_gorm

import (
	"gorm.io/gorm"

	"gitlab.com/nunet/device-management-service/internal/repositories"
	"gitlab.com/nunet/device-management-service/models"
)

// VirtualMachineRepositoryGORM is a GORM implementation of the VirtualMachineRepository interface.
type VirtualMachineRepositoryGORM struct {
	repositories.GenericRepository[models.VirtualMachine]
}

// NewVirtualMachineRepository creates a new instance of VirtualMachineRepositoryGORM.
// It initializes and returns a GORM-based repository for VirtualMachine entities.
func NewVirtualMachineRepository(db *gorm.DB) repositories.VirtualMachineRepository {
	return &VirtualMachineRepositoryGORM{
		NewGenericRepository[models.VirtualMachine](db),
	}
}
