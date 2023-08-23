package db

import (
	"fmt"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gitlab.com/nunet/device-management-service/internal/config"
	"gitlab.com/nunet/device-management-service/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

type DMSDB interface {
	Create(value interface{}) error
	Save(value interface{}) (int64, error)
	Delete(value interface{}) (int64, error)
	Find(dest interface{}) (int64, error)
	WhereFind(dest interface{}, attrib string, value string) error
	WhereFirst(dest interface{}, attrib string, value string) error
	Updates(model interface{}, dest interface{}, attrib string, value string, values interface{}) error
}

type DMSGormDB struct {
	db *gorm.DB
}

func (db *DMSGormDB) ConnectDatabase(path string) {
	database, err := gorm.Open(sqlite.Open(path), &gorm.Config{})

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

	db.db = database
	if err := DB.Use(otelgorm.NewPlugin()); err != nil {
		panic(err)
	}
}

func (g *DMSGormDB) Create(value interface{}) error {
	result := g.db.Create(&value)
	if result.Error != nil {
		return fmt.Errorf("couldn't create data in database: %v", result.Error)
	}
	return nil
}

func (g *DMSGormDB) Save(value interface{}) (int64, error) {
	result := g.db.Save(&value)
	if result.Error != nil {
		return 0, fmt.Errorf("couldn't find data in database: %v", result.Error)
	}
	return result.RowsAffected, nil
}

func (g *DMSGormDB) Delete(value interface{}) (int64, error) {
	result := g.db.Delete(&value)
	if result.Error != nil {
		return 0, fmt.Errorf("couldn't find data in database: %v", result.Error)
	}
	return result.RowsAffected, nil
}

func (g *DMSGormDB) Find(dest interface{}) (int64, error) {
	result := g.db.Find(&dest)
	if result.Error != nil {
		return 0, fmt.Errorf("couldn't find data in database: %v", result.Error)
	}
	return result.RowsAffected, nil
}

func (g *DMSGormDB) WhereFind(dest interface{}, attrib string, value string) error {
	result := g.db.Where(attrib+" = ?", value).Find(&dest)
	if result.Error != nil {
		return fmt.Errorf("Couldn't read "+attrib+" From DB: %v", result.Error)
	}
	return nil
}

func (g *DMSGormDB) WhereFirst(dest interface{}, attrib string, value string) error {
	result := g.db.Where(attrib+" = ?", value).First(&dest)
	if result.Error != nil {
		return fmt.Errorf("Couldn't read "+attrib+" From DB: %v", result.Error)
	}
	return nil
}

func (g *DMSGormDB) Updates(model interface{}, dest interface{}, attrib string, value string, values interface{}) error {
	result := DB.Model(&model).Where(attrib+" = ?", value).Updates(values)
	if result.Error != nil {
		return fmt.Errorf("couldn't update data in database: %v", result.Error)
	}
	return nil
}

func ConnectDatabase() {
	database, err := gorm.Open(sqlite.Open(fmt.Sprintf("%s/nunet.db", config.GetConfig().General.MetadataPath)), &gorm.Config{})

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
