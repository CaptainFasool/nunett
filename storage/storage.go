// Package storage handles mainly storage access to remote storage providers such as
// AWS S3 and IPFS. Also, it handles the control of storage volumes.
package storage

import (
	"gitlab.com/nunet/device-management-service/models"
)

type MiB uint64

// StorageProvider handles I/O operations of files with remote storage providers
// such as AWS S3 and IPFS.
//
// Its functionality is coupled with local mounted volumes, meaning that implementations
// will rely on mounted files to upload data and downloading data will result in a
// local mounted volume.
//
// - When needed, the availability-checking (e.g.: check if IPFS node is running) of some
// storage provider is handled when instantiating the implementation
//
// - Any necessary authentication data is also provided within the `*models.SpecConfig`
// parameters
//
// - Although it may be feasible to implement, the interface was not built with
// the idea of supporting streaming of data and non-file storage operations (e.g.:
// some databases)
type StorageProvider interface {
	// Upload uploads a storage volume data to a given remote storage provider.
	// The operation results in a return value that might vary from provider to provider
	// (and it may not exist in some cases).
	Upload(vol StorageVolume, target *models.SpecConfig) (*models.SpecConfig, error)

	// Download downloads data from a given source, mounting it to a certain local path.
	// (Note: the operation will fail if the user running DMS does not have access permission
	// to the given path).
	Download(source *models.SpecConfig, outputPath string) (StorageVolume, error)

	// Size returns the size of a given source. The method may also be useful to check
	// if a given source is available.
	Size(source *models.SpecConfig) (MiB, error)
}

// StorageVolume contains mainly the path to a directory/file where data is mounted
// and additional metadata which is not yet defined.
type StorageVolume struct {
	// Path is the path to a directory/file where data is mounted
	Path string
}

// VolumeController is used to manage storage volumes which are data mounted to files/directories.
type VolumeController interface {
	CreateVolume(path string, volType string) (StorageVolume, error)
	DeleteVolume(in StorageVolume) error
	ListVolumes() ([]StorageVolume, error)
}
