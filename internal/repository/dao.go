package repository

import (
	"fmt"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DAO interface {
	NewServiceQuery() ServiceQuery
	NewResourcesQuery() ResourcesQuery
}

type dao struct{}

var DB *gorm.DB

func NewDAO() DAO {
	return &dao{}
}

func NewDB() (*gorm.DB, error) {
	database, err := gorm.Open(sqlite.Open(fmt.Sprintf("%s/nunet.db", config.GetConfig().General.MetadataPath)), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database!")
	}

	database.AutoMigrate(&models.VirtualMachine{})
	database.AutoMigrate(&models.Machine{})
	database.AutoMigrate(&models.AvailableResources{})
	database.AutoMigrate(&models.FreeResources{})
	database.AutoMigrate(&models.PeerInfo{})
	database.AutoMigrate(&models.Services{})
	database.AutoMigrate(&models.ServiceResourceRequirements{})
	database.AutoMigrate(&models.RequestTracker{})
	database.AutoMigrate(&models.Libp2pInfo{})
	database.AutoMigrate(&models.DeploymentRequestFlat{})
	database.AutoMigrate(&models.MachineUUID{})
	database.AutoMigrate(&models.Connection{})

	DB = database
	if err := DB.Use(otelgorm.NewPlugin()); err != nil {
		panic(err)
	}
	return DB, nil
}

func (d *dao) NewServiceQuery() ServiceQuery {
	return &serviceQuery{}
}

func (d *dao) NewResourcesQuery() ResourcesQuery {
	return &resourceQuery{}
}
