package db

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase(fs afero.Fs) {

	dbPath := fmt.Sprintf("%s/nunet.db", config.GetConfig().General.MetadataPath)

	// Check if the database file exists in the provided filesystem, and if not, create it
	if exists, _ := afero.Exists(fs, dbPath); !exists {
		afero.WriteFile(fs, dbPath, []byte{}, 0644)
	}

	// Open the SQLite database using the path from the filesystem
	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database!")
	}
	database.AutoMigrate(&models.ElasticToken{})
	database.AutoMigrate(&models.VirtualMachine{})
	database.AutoMigrate(&models.Machine{})
	database.AutoMigrate(&models.AvailableResources{})
	database.AutoMigrate(&models.FreeResources{})
	database.AutoMigrate(&models.PeerInfo{})
	database.AutoMigrate(&models.Services{})
	database.AutoMigrate(&models.ServiceResourceRequirements{})
	database.AutoMigrate(&models.ContainerImages{})
	database.AutoMigrate(&models.RequestTracker{})
	database.AutoMigrate(&models.Libp2pInfo{})
	database.AutoMigrate(&models.DeploymentRequestFlat{})
	database.AutoMigrate(&models.MachineUUID{})
	database.AutoMigrate(&models.Connection{})
	database.AutoMigrate(&models.LogBinAuth{})

	DB = database
	if err := DB.Use(otelgorm.NewPlugin()); err != nil {
		panic(err)
	}
}
