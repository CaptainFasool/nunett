package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/storage"
)

type VolumeControllerTestSuite struct {
	suite.Suite
	volController *BasicVolumeController
	volumes       map[string]*storage.StorageVolume
	fs            afero.Fs
}

func TestVolumeControllerTestSuite(t *testing.T) {
	suite.Run(t, new(VolumeControllerTestSuite))
}

func (s *VolumeControllerTestSuite) SetupTest() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(s.T(), err)

	basePath := "/home/.nunet/volumes/"
	fs := afero.NewMemMapFs()

	err = fs.MkdirAll(basePath, 0755)
	assert.NoError(s.T(), err)

	s.fs = fs
	s.volController, err = NewDefaultVolumeController(db, basePath, fs)
	assert.NoError(s.T(), err)

	s.volumes = map[string]*storage.StorageVolume{
		"volume1": {
			Path:           basePath + "volume1",
			ReadOnly:       false,
			Private:        false,
			EncryptionType: models.EncryptionTypeNull,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		"volume2": {
			CID:            "baf222",
			Path:           basePath + "volume2",
			ReadOnly:       false,
			Private:        false,
			EncryptionType: models.EncryptionTypeNull,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	for _, vol := range s.volumes {
		// create root volume dir
		err = fs.MkdirAll(vol.Path, 0755)
		assert.NoError(s.T(), err)

		// create volume record in db
		err = db.Create(vol).Error
		assert.NoError(s.T(), err)
	}

	// Write a file in volume2
	err = afero.WriteFile(fs, s.volumes["volume2"].Path+"/file.txt", []byte("hello world"), 0644)
	assert.NoError(s.T(), err)
}

func (s *VolumeControllerTestSuite) TearDownTest() {
	// Clean up the test environment
	err := s.volController.db.Exec("DELETE FROM storage_volumes").Error
	assert.NoError(s.T(), err)
	s.volController.db = nil
	s.volController = nil
	s.fs = nil
}

func (s *VolumeControllerTestSuite) TestCreateVolume() {
	// Test case 1: Create a volume without options
	vol1, err := s.volController.CreateVolume(storage.VolumeSourceS3)
	assert.NoError(s.T(), err)

	// Test case 2: Create a volume with private option
	vol2, err := s.volController.CreateVolume(storage.VolumeSourceS3, WithPrivate[storage.CreateVolOpt]())
	assert.NoError(s.T(), err)

	// Verify returned volume details for test case 1
	assert.NotEmpty(s.T(), vol1.Path)
	assert.Empty(s.T(), vol1.CID)
	assert.Equal(s.T(), false, vol1.Private)
	assert.Equal(s.T(), false, vol1.ReadOnly)
	assert.Equal(s.T(), models.EncryptionTypeNull, vol1.EncryptionType)

	// Verify returned volume details for test case 2
	assert.NotEmpty(s.T(), vol2.Path)
	assert.Empty(s.T(), vol2.CID)
	assert.Equal(s.T(), true, vol2.Private)
	assert.Equal(s.T(), false, vol2.ReadOnly)
	assert.Equal(s.T(), models.EncryptionTypeNull, vol2.EncryptionType)

	// Verify that the volumes are stored in the database
	var volumes []storage.StorageVolume
	err = s.volController.db.Find(&volumes).Error
	assert.NoError(s.T(), err)
	assert.Len(s.T(), volumes, 4) // there are already 2 volumes created in the suite
	// TODO-maybe: should we also check the DB content for each volume?

	// check if directories were created based on volumes path
	fileInfoVol1, err := s.fs.Stat(vol1.Path)
	assert.NoError(s.T(), err)
	assert.True(s.T(), fileInfoVol1.IsDir())

	fileInfoVol2, err := s.fs.Stat(vol2.Path)
	assert.NoError(s.T(), err)
	assert.True(s.T(), fileInfoVol2.IsDir())
}

func (s *VolumeControllerTestSuite) TestLockVolume() {
	testCases := []struct {
		name        string
		volumePath  string
		cid         string
		private     bool
		expectError bool
	}{
		{
			name:        "Lock volume with CID",
			volumePath:  s.volumes["volume1"].Path,
			cid:         "abcdef",
			private:     false,
			expectError: false,
		},
		{
			name:        "Lock volume with private option",
			volumePath:  s.volumes["volume2"].Path,
			cid:         s.volumes["volume2"].CID,
			private:     true,
			expectError: false,
		},
		{
			name:        "Lock non-existent volume",
			volumePath:  "/path/to/non-existent-volume",
			cid:         "",
			private:     false,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			var opts []storage.LockVolOpt
			if tc.cid != "" {
				opts = append(opts, WithCID(tc.cid))
			}
			if tc.private {
				opts = append(opts, WithPrivate[storage.LockVolOpt]())
			}

			err := s.volController.LockVolume(tc.volumePath, opts...)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// verify database fields (readOnly must be true)
				var vol storage.StorageVolume
				err = s.volController.db.Where("path = ?", tc.volumePath).First(&vol).Error
				assert.NoError(t, err)
				assert.True(t, vol.ReadOnly)
				// checking CID and Private as some test cases inputed them
				assert.Equal(t, tc.cid, vol.CID)
				assert.Equal(t, tc.private, vol.Private)

				// verifying volume dir is read-only
				fileInfo, err := s.fs.Stat(tc.volumePath)
				assert.NoError(t, err)
				assert.Equal(t, os.FileMode(0400), fileInfo.Mode().Perm())
			}
		})
	}
}

func (s *VolumeControllerTestSuite) TestDeleteVolume() {
	testCases := []struct {
		name           string
		identifier     string
		identifierType storage.IDType
		expectError    bool
	}{
		{
			name:           "Delete volume by path",
			identifier:     s.volumes["volume1"].Path,
			identifierType: storage.IDTypePath,
			expectError:    false,
		},
		{
			name:           "Delete volume by CID",
			identifier:     s.volumes["volume2"].CID,
			identifierType: storage.IDTypeCID,
			expectError:    false,
		},
		{
			name:           "Delete non-existent volume by path",
			identifier:     "/path/to/non-existent-volume",
			identifierType: storage.IDTypePath,
			expectError:    true,
		},
		{
			name:           "Delete non-existent volume by CID",
			identifier:     "non-existent-cid",
			identifierType: storage.IDTypeCID,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		err := s.volController.DeleteVolume(tc.identifier, tc.identifierType)
		if tc.expectError {
			assert.Error(s.T(), err)
		} else {
			assert.NoError(s.T(), err)
		}
	}

	// Verify that the volumes are deleted from the database
	var volumes []storage.StorageVolume
	err := s.volController.db.Find(&volumes).Error
	assert.NoError(s.T(), err)
	assert.Len(s.T(), volumes, 0)
}

func (s *VolumeControllerTestSuite) TestListVolumes() {
	volumes, err := s.volController.ListVolumes()
	assert.NoError(s.T(), err)

	assert.Len(s.T(), volumes, len(s.volumes))

	// assert details of returned volumes
	for _, retVol := range volumes {
		// Find the corresponding volume in the test suite's volumes map
		expectedVol, ok := s.volumes[filepath.Base(retVol.Path)]
		assert.True(s.T(), ok, "Unexpected volume returned: %s", retVol.Path)

		// Assert the properties of the returned volume
		assert.Equal(s.T(), expectedVol.CID, retVol.CID)
		assert.Equal(s.T(), expectedVol.Path, retVol.Path)
		assert.Equal(s.T(), expectedVol.ReadOnly, retVol.ReadOnly)
		assert.Equal(s.T(), expectedVol.Private, retVol.Private)
		assert.Equal(s.T(), expectedVol.EncryptionType, retVol.EncryptionType)
		assert.True(s.T(), expectedVol.CreatedAt.Equal(retVol.CreatedAt))
		assert.True(s.T(), expectedVol.UpdatedAt.Equal(retVol.UpdatedAt))
	}
}

func (s *VolumeControllerTestSuite) TestGetSize() {
	size, err := s.volController.GetSize(s.volumes["volume1"].Path, storage.IDTypePath)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(0), size)
}
