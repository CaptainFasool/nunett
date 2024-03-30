package repositories

import (
	"gitlab.com/nunet/device-management-service/models"
)

// PeerInfoRepository represents a repository for CRUD operations on PeerInfo entities.
type PeerInfoRepository interface {
	GenericRepository[models.PeerInfo]
}

// MachineRepository represents a repository for CRUD operations on Machine entities.
type MachineRepository interface {
	GenericRepository[models.Machine]
}

// FreeResourcesRepository represents a repository for CRUD operations on FreeResources entity.
type FreeResourcesRepository interface {
	GenericEntityRepository[models.FreeResources]
}

// AvailableResourcesRepository represents a repository for CRUD operations on AvailableResources entity.
type AvailableResourcesRepository interface {
	GenericEntityRepository[models.AvailableResources]
}

// ServicesRepository represents a repository for CRUD operations on Services entities.
type ServicesRepository interface {
	GenericRepository[models.Services]
}

// ServiceResourceRequirementsRepository represents a repository for CRUD operations on ServiceResourceRequirements entities.
type ServiceResourceRequirementsRepository interface {
	GenericRepository[models.ServiceResourceRequirements]
}

// Libp2pInfoRepository represents a repository for CRUD operations on Libp2pInfo entity.
type Libp2pInfoRepository interface {
	GenericEntityRepository[models.Libp2pInfo]
}

// MachineUUIDRepository represents a repository for CRUD operations on MachineUUID entity.
type MachineUUIDRepository interface {
	GenericEntityRepository[models.MachineUUID]
}

// ConnectionRepository represents a repository for CRUD operations on Connection entities.
type ConnectionRepository interface {
	GenericRepository[models.Connection]
}

// ElasticTokenRepository represents a repository for CRUD operations on ElasticToken entities.
type ElasticTokenRepository interface {
	GenericRepository[models.ElasticToken]
}
