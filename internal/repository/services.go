package repository

import (
	"gitlab.com/nunet/device-management-service/db"
	"gitlab.com/nunet/device-management-service/models"
)

type ServiceQuery interface {
	CreateService(service models.Services) (*int64, error)
	GetService(id int64) (*models.Services, error)
	DeleteService(serviceID int64) error
	UpdateService(service models.Services) (*models.Services, error)
}

type serviceQuery struct{}

func (u *serviceQuery) CreateService(service models.Services) (*int64, error) {

	if err := db.DB.Create(&service).Error; err != nil {
		// zlog.Sugar().Errorf("couldn't save service: %v", err)
		return nil, err
	}
	var id int64
	return &id, nil
}

func (u *serviceQuery) GetService(id int64) (*models.Services, error) {
	// TODO: implement logic for getservice

	service := models.Services{}

	return &service, nil
}

func (u *serviceQuery) DeleteService(serviceID int64) error {
	// TODO: implement logic for deleteservice
	return nil
}

func (u *serviceQuery) UpdateService(service models.Services) (*models.Services, error) {

	var updatedService models.Services

	return &updatedService, nil
}
