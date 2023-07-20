package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	database, err := gorm.Open(sqlite.Open(fmt.Sprintf("%s/nunet.db", config.GetConfig().General.MetadataPath)), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database!")
	}

	database.AutoMigrate(&models.VirtualMachine{})
	database.AutoMigrate(&models.Machine{})
	database.AutoMigrate(&models.AvailableResources{})
	// database.AutoMigrate(&models.FreeResources{})
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

	sqlDB, err := DB.DB()
	if err != nil {
		zlog.Sugar().Errorf("Failed to get *sql.DB from gorm.DB: %v", err)
		panic(err)
	}

	tables := []string{"free_resources"}
	for _, table := range tables {
		err = applySQLTableMigrations(sqlDB, table)
		if err != nil {
			panic(err)
		}
	}
}

func applySQLTableMigrations(sqlDB *sql.DB, table string) error {
	driver, err := sqlite3.WithInstance(sqlDB, &sqlite3.Config{})
	if err != nil {
		zlog.Sugar().Errorf("Failed to prepare database driver: %v", err)
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://db/migrations/%s", table),
		"sqlite3", driver)
	if err != nil {
		zlog.Sugar().Errorf("Failed to prepare migration: %v", err)
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		zlog.Sugar().Errorf("Failed to apply migrations: %v", err)
		return err
	}
	return nil
}
