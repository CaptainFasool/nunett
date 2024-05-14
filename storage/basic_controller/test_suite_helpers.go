package basic_controller

import (
	"fmt"

	"github.com/spf13/afero"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gitlab.com/nunet/device-management-service/storage"
	"gitlab.com/nunet/device-management-service/utils"
)

type VolControllerTestSuiteHelper struct {
	BasicVolController *BasicVolumeController
	Fs                 afero.Fs
	Db                 *gorm.DB
	Volumes            map[string]*storage.StorageVolume
}

// SetupVolControllerTestSuite sets up a volume controller with 0-n volumes given a base path.
// If volumes are inputed, directories will be created and volumes will be stored in the database
func SetupVolControllerTestSuite(basePath string,
	volumes map[string]*storage.StorageVolume) (*VolControllerTestSuiteHelper, error) {

	randomStr := utils.RandomString(16)
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", randomStr)), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	fs := afero.NewMemMapFs()

	err = fs.MkdirAll(basePath, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	vc, err := NewDefaultVolumeController(db, basePath, fs)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume controller: %w", err)
	}

	for _, vol := range volumes {
		// create root volume dir
		err = fs.MkdirAll(vol.Path, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create volume dir: %w", err)
		}

		// create volume record in db
		err = db.Create(vol).Error
		if err != nil {
			return nil, fmt.Errorf("failed to create volume record: %w", err)
		}
	}

	return &VolControllerTestSuiteHelper{vc, fs, db, volumes}, nil
}
