package db

import (
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gitlab.com/nunet/device-management-service/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	database, err := gorm.Open(sqlite.Open("/etc/nunet/nunet.db"), &gorm.Config{})

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
}
