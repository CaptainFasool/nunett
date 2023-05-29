package repository

import (
	"gitlab.com/nunet/device-management-service/models"
)

type ResourcesQuery interface {
	CreateFreeResources(freeResource models.FreeResources) (*int64, error)
	GetFreeResources() (*models.FreeResources, error)
	UpdateFreeResources(freeResource models.FreeResources) (*models.FreeResources, error)
	CreateAvailableResources(availableResource models.AvailableResources) (*int64, error)
	GetAvailableResources() (*models.AvailableResources, error)
	UpdateAvailableResources(availbleResource models.AvailableResources) (*models.AvailableResources, error)
}

type resourceQuery struct{}

func (u *resourceQuery) CreateFreeResources(freeResource models.FreeResources) (*int64, error) {
	// TODO: implement logic for createfreeresources

	var id int64
	return &id, nil
}

func (u *resourceQuery) GetFreeResources() (*models.FreeResources, error) {
	// TODO: implement logic for getservice

	freeResource := models.FreeResources{}

	return &freeResource, nil
}

func (u *resourceQuery) UpdateFreeResources(freeResource models.FreeResources) (*models.FreeResources, error) {
	// TODO: implement logic for updatefreeresources

	var updatedFreeResources models.FreeResources

	return &updatedFreeResources, nil
}

func (u *resourceQuery) CreateAvailableResources(availbleResource models.AvailableResources) (*int64, error) {
	// TODO: implement logic for createavailableresources

	var id int64
	return &id, nil
}

func (u *resourceQuery) GetAvailableResources() (*models.AvailableResources, error) {
	// TODO: implement logic for getservice

	availableResource := models.AvailableResources{}

	return &availableResource, nil
}

func (u *resourceQuery) UpdateAvailableResources(freeResource models.AvailableResources) (*models.AvailableResources, error) {
	// TODO: implement logic for updatefreeresources

	var updatedAvailableResources models.AvailableResources

	return &updatedAvailableResources, nil
}
