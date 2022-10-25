package db

import (
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

	DB = database
}
