package storage

import (
	"fmt"
	"time"

	"github.com/spf13/afero"
	"gorm.io/gorm"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/storage"
	"gitlab.com/nunet/device-management-service/utils"
)

// BasicVolumeController is the default implementation of the VolumeController.
// It persists storage volumes information in the local database.
//
// TODO [Remove gormDB]: *gorm.DB shouldn't be coupled with this struct.
// We should otherwise depend on a DB general interface
// which we may probably have with the DB refactoring
type BasicVolumeController struct {
	// db is where all volumes information is stored.
	db *gorm.DB

	// basePath is the base path where volumes are stored under
	basePath string

	// file system to act upon
	fs afero.Fs
}

// NewDefaultVolumeController returns a new instance of BasicVolumeController
//
// TODO-BugFix [path]: volBasePath might not end with `/`, causing errors when calling methods.
// We need to validate it using the `path` library or just verifying the string.
func NewDefaultVolumeController(db *gorm.DB, volBasePath string, fs afero.Fs) (*BasicVolumeController, error) {
	// TODO: I'm not sure how the automigration will be placed on the new refactoring.
	// Let's keep here until the database refactoring is done
	err := db.AutoMigrate(&storage.StorageVolume{})
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate storage volumes: %w", err)
	}

	return &BasicVolumeController{
		db:       db,
		basePath: volBasePath,
		fs:       fs,
	}, nil
}

// CreateVolume creates a new storage volume given a source (S3, IPFS, job, etc). The
// creation of a storage volume effectively creates an empty directory in the local filesystem
// and writes a record in the database.
//
// The directory name follows the format: `<volSource> + "-" + <name>
// where `name` is random.
//
// TODO-maybe [withName]: allow callers to specify custom name for path
func (vc *BasicVolumeController) CreateVolume(volSource storage.VolumeSource, opts ...storage.CreateVolOpt) (storage.StorageVolume, error) {
	vol := &storage.StorageVolume{
		Private:        false,
		ReadOnly:       false,
		EncryptionType: models.EncryptionTypeNull,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	for _, opt := range opts {
		opt(vol)
	}

	vol.Path = vc.basePath + string(volSource) + "-" + utils.RandomString(16)

	if err := vc.fs.Mkdir(vol.Path, 0770); err != nil {
		return storage.StorageVolume{}, fmt.Errorf("failed to create storage volume: %w", err)
	}

	if err := vc.db.Create(vol).Error; err != nil {
		return storage.StorageVolume{}, err
	}

	return *vol, nil
}

// LockVolume makes the volume read-only, not only changing the field value but also changing file permissions.
// It should be used after all necessary data has been written.
// It optionally can also set the CID and mark the volume as private.
//
// TODO-maybe [CID]: maybe calculate CID of every volume in case WithCID opt is not provided
func (vc *BasicVolumeController) LockVolume(pathToVol string, opts ...storage.LockVolOpt) error {
	var vol storage.StorageVolume
	if err := vc.db.Where("path = ?", pathToVol).First(&vol).Error; err != nil {
		return fmt.Errorf("failed to find storage volume with path %s - Error: %w", pathToVol, err)
	}

	for _, opt := range opts {
		opt(&vol)
	}

	// update records
	vol.ReadOnly = true
	vol.UpdatedAt = time.Now()
	if err := vc.db.Where("path = ?", pathToVol).Save(&vol).Error; err != nil {
		return fmt.Errorf("failed to update storage volume with path %s - Error: %w", pathToVol, err)
	}

	// change file permissions
	if err := vc.fs.Chmod(vol.Path, 0400); err != nil {
		return fmt.Errorf("failed to make storage volume read-only (path: %s): %w", vol.Path, err)
	}

	return nil
}

// WithPrivate designates a given volume as private. It can be used both
// when creating or locking a volume.
func WithPrivate[T storage.CreateVolOpt | storage.LockVolOpt]() T {
	return func(v *storage.StorageVolume) {
		v.Private = true
	}
}

// WithCID sets the CID of a given volume if already calculated
//
// TODO [validate]: check if CID provided is valid
func WithCID(cid string) storage.LockVolOpt {
	return func(v *storage.StorageVolume) {
		v.CID = cid
	}
}

// DeleteVolume deletes a given storage volume record from the database.
// Identifier is either a CID or a path of a volume. Therefore, records for both
// will be deleted.
//
// Note [CID]: if we start to type CID as cid.CID, we may have to use generics here
// as in `[T string | cid.CID]`
func (vc *BasicVolumeController) DeleteVolume(identifier string, idType storage.IDType) error {

	var result *gorm.DB

	switch idType {
	case storage.IDTypePath:
		result = vc.db.Where("path = ?", identifier).Delete(&storage.StorageVolume{})
	case storage.IDTypeCID:
		// TODO: c_id because gorm automatically does the transformation. I didn't put gorm tags
		// because I think it's better to wait for DB refactoring
		result = vc.db.Where("c_id = ?", identifier).Delete(&storage.StorageVolume{})
	default:
		return fmt.Errorf("identifier type not supported")
	}

	if result.Error != nil {
		return fmt.Errorf("failed to delete volume: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("volume not found")
	}

	return nil
}

// ListVolumes returns a list of all storage volumes stored on the database
//
// TODO [filter]: maybe add opts to filter results by certain values
func (vc *BasicVolumeController) ListVolumes() ([]storage.StorageVolume, error) {
	var volumes []storage.StorageVolume
	err := vc.db.Find(&volumes).Error
	if err != nil {
		return nil, err
	}
	return volumes, nil
}

// GetSize returns the size of a volume
// TODO-minor: identify which measurement type will be used
func (vc *BasicVolumeController) GetSize(identifier string, idType storage.IDType) (int64, error) {
	var path string

	switch idType {
	case storage.IDTypePath:
		path = identifier
	case storage.IDTypeCID:
		var vol storage.StorageVolume
		if err := vc.db.Where("cid = ?", identifier).First(&vol).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return 0, fmt.Errorf("volume with CID %s not found", identifier)
			}
			return 0, fmt.Errorf("failed to query volume by CID: %w", err)
		}
		path = vol.Path
	default:
		return 0, fmt.Errorf("unsupported ID type: %d", idType)
	}

	size, err := utils.GetDirectorySize(vc.fs, path)
	if err != nil {
		return 0, fmt.Errorf("failed to get directory size: %w", err)
	}

	return size, nil
}

// EncryptVolume encrypts a given volume
func (vc *BasicVolumeController) EncryptVolume(path string, encryptor models.Encryptor, encryptionType models.EncryptionType) error {
	return fmt.Errorf("not implemented")
}

// DecryptVolume decrypts a given volume
func (vc *BasicVolumeController) DecryptVolume(path string, decryptor models.Decryptor, decryptionType models.EncryptionType) error {
	return fmt.Errorf("not implemented")
}

// TODO-minor: compiler-time check for interface implementation
var _ storage.VolumeController = (*BasicVolumeController)(nil)
