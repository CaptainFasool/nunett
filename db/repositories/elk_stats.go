package repositories

import (
	"gitlab.com/nunet/device-management-service/models"
)

// RequestTrackerRepository represents a repository for CRUD operations on RequestTracker entities.
type RequestTrackerRepository interface {
	GenericRepository[models.RequestTracker]
}
