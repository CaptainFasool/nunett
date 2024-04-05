package storage

/*

Comments:
1. This file considers storage interfaces dealing directly with volumes.
Therefore, it's coupled with the file system.
2. VolumeController is used by all versions

TODOs and questions:
1. How to handle different storage provider params?
    1.1. SpecConfig + InputSource or
    1.2. Reader/Writer instantiated with different params for each implementation

*/

// StorageVolume is basically the path to a directory/file where data is contained.
// Plus some metadata.
type StorageVolume struct {
	Path            string
	TimeLastUpdated int64
}

// VolumeController is used to manage storage volumes. It'll probably be used
// by storage provider interfaces.
type VolumeController interface {
	CreateVolume(path string, volType string) (StorageVolume, error)
	DeleteVolume(in StorageVolume) error
	ListVolumes() ([]StorageVolume, error)
	GetCapabilities(in StorageVolume) (map[string]interface{}, error)
	GetVolumeStats(in StorageVolume) (map[string]interface{}, error)
	CreateSnapshot(in StorageVolume) (StorageVolume, error)
	DeleteSnapshot(snapshotID string) error
	ListSnapshots() ([]StorageVolume, error)
}

//= = = = = = = = = =  VERSION 1: InputSource/OutputTarget  = = = = = = = = = =

/*

Pros:
- It may be easier to understand the code at first. StorageProvider interface is not
hard to understand or use it.
- One unique implementation for both upload/download operations and other secundary operations

Cons:
- It does not have compile-time type checking as it relies on `map[string]interfaces{}` to be
decoded by the caller based on a given type (as executor is doing)
- Although the interface is easier to understand, having this general type such as InputSource/OutputTarget
makes the params harder to understand because they're not directly coupled with the interface. Each implementation
will use their own decoder.
*/

// InputSource/StoreResult (temporary) will be used by different storage type implementations
// which need different params when reading/writing from/to them.
// (similar to SpecConfig used by the Executor)
type InputSource map[string]interface{}
type OutputTarget map[string]interface{}

// StoreResult as described above + contains metadata returned by the store operation
type StoreResult map[string]interface{}

// StorageProvider
//
// - Authentication done when instantiating the implementation
//   - CONCERN: different files might have different auth within the same account
//
// - IsAvailable() checked when instantiating the implementation
type StorageProvider interface {
	// Upload uploads the data to the source based on StorageVolume and
	// the target of operation
	//
	// CONCERN: do we really need to get returned StoreResult?
	// I'm afraid we might need to get any metadata returned after store operation
	// for some cases
	Upload(input StorageVolume, target OutputTarget) (StoreResult, error)

	// Download downloads the data from the source
	Download(input InputSource, outputPath string) (StorageVolume, error)

	// Size is also used to check if it exists
	Size(identifier string) (int64, error)
}

//= = = = = = = = = =  VERSION 2: StorageReader/Writer  = = = = = = = = = =

/*

This version eliminates the need of the InputSource/OutputTarget.

Pros:
- Compile-time type checking
- Modular interfaces

Cons:
- It may be harder to understand at first.
- Unusual way of dealing with storage provider maybe?

*/

// StorageReader: when instantiating the implementation, the reader will contain
// information regarding necessary params to read from the source such as: key,
// paths, auth, etc.
type StorageReader interface {
	// Download downloads the data from the source
	Download(outputPath string) (StorageVolume, error)

	// Size is also used to check if it exists
	Size(identifier string) (int64, error)
}

// StorageWriter: when instantiating the implementation, the writer will contain
// information regarding necessary params to write to the source such as: key,
// paths, auth, etc.
type StorageWriter interface {
	// Upload uploads the data to the source based on the information inputed when
	// the writer was instantiated
	Upload(in StorageVolume) error
}
