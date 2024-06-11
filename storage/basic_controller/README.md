# Introduction

This sub package offers a default implementation of the volume controller.

# Stucture and organisation

Here is quick overview of the contents of this pacakge:

* [README](https://gitlab.com/nunet/device-management-service/-/blob/develop/storage/basic_controller/README.md): Current file which is aimed towards developers who wish to use and modify the functionality.

* [basic_controller](https://gitlab.com/nunet/device-management-service/-/blob/develop/storage/basic_controller/basic_controller.go): This file implements the methods for `VolumeController` interface.

* [basic_controller_test](https://gitlab.com/nunet/device-management-service/-/blob/develop/storage/basic_controller/basic_controller_test.go): This file contains the unit tests for the methods of `VolumeController` interface.

# Contributing

For guidelines of how to contribute, install and test the `device-management-service` component which contains `storage` package, please refer to package level documentation:

* Storage package level [../README.md](https://gitlab.com/nunet/device-management-service/-/blob/develop/storage/README.md)
* DMS component level [../../README.md](https://gitlab.com/nunet/device-management-service/-/blob/develop/README.md)
* Contribution guidelines [../../CONTRIBUTING.md](https://gitlab.com/nunet/device-management-service/-/blob/develop/CONTRIBUTING.md)
* Code of conduct [../../CODE_OF_CONDUCT.md](https://gitlab.com/nunet/device-management-service/-/blob/develop/CODE_OF_CONDUCT.md)
* [Secure coding guidelines](https://gitlab.com/nunet/documentation/-/wikis/secure-coding-guidelines)

# Specifications overview

Refer to the [specifications overview](https://gitlab.com/nunet/device-management-service/-/blob/develop/storage/README.md#specifications-overview) in the package readme.

# Functions

### NewDefaultVolumeController

* signature: `NewDefaultVolumeController(db *gorm.DB, volBasePath string, fs afero.Fs) -> (storage.basic_controller.BasicVolumeController, error)` <br/>
* input #1: local database instance of type `*gorm.DB` <br/>
* input #2: base path of the volumes <br/>
* input #3: file system instance of type `afero.FS` <br/>
* output (sucess): new instance of type `BasicVolumeController` <br/>
* output (error): error

`NewDefaultVolumeController` returns a new instance of `BasicVolumeController` struct. 

`BasicVolumeController` is the default implementation of the `VolumeController` interface. It persists storage volumes information in the local database.

### CreateVolume

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/storage/README.md#createvolume)

`CreateVolume` creates a new storage volume given a storage source (S3, IPFS, job, etc). The creation of a storage volume effectively creates an empty directory in the local filesystem and writes a record in the database.

The directory name follows the format: `<volSource> + "-" + <name>`  where `name` is random.

`CreateVolume` will return an error if there is a failure in
* creation of new directory
* creating a database entry

See [Feature: Creating a storage volume](https://gitlab.com/nunet/test-suite/-/blob/develop/stages/functional_tests/features/device-management-service/storage/basic_controller/Create_Volume.feature) for test cases including error scenarios.

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/storage/basic_controller/sequences/createVolume.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/storage/basic_controller/sequences/rendered/createVolume.sequence.svg)

### LockVolume

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/storage/README.md#lockvolume)

`LockVolume` makes the volume read-only, not only changing the field value but also changing file permissions. It should be used after all necessary data has been written to the volume. It optionally can also set the CID and mark the volume as private

`LockVolume` will return an error when
* No storage volume is found at the specified
* There is error in saving the updated volume in the database
* There is error in updating file persmissions

See [Feature: Locking a storage volume](https://gitlab.com/nunet/test-suite/-/blob/develop/stages/functional_tests/features/device-management-service/storage/basic_controller/Lock_Volume.feature)

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/storage/basic_controller/sequences/lockVolume.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/storage/basic_controller/sequences/rendered/lockVolume.sequence.svg)

### DeleteVolume

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/storage/README.md#deletevolume)

`DeleteVolume` deletes a given storage volume record from the database. The identifier can be a path of a volume or a Content ID (CID). Therefore, records for both will be deleted.

It will return an error when
* Input has incorrect identifier
* There is failure in deleting the volume
* No volume is found

See [Feature: Deleting a storage volume](https://gitlab.com/nunet/test-suite/-/blob/develop/stages/functional_tests/features/device-management-service/storage/basic_controller/Delete_Volume.feature)

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/storage/basic_controller/sequences/deleteVolume.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/storage/basic_controller/sequences/rendered/deleteVolume.sequence.svg)

### ListVolumes

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/storage/README.md#listvolumes)

`ListVolumes` function returns a list of all storage volumes stored on the database.

It will return an error when no storage volumes exist.

See [Feature: Listing storage volumes](https://gitlab.com/nunet/test-suite/-/blob/develop/stages/functional_tests/features/device-management-service/storage/basic_controller/List_Volumes.feature)

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/storage/basic_controller/sequences/listVolumes.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/storage/basic_controller/sequences/rendered/listVolumes.sequence.svg)

### GetSize

For function signature refer to the package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/storage/README.md#getsize)

`GetSize` returns the size of a volume. The input can be a path or a Content ID (CID).

It will return an error if the operation fails due to:
* error while querying database
* volume not found for given identifier
* unsupported identifed provided as input
* error while caculating size of directory

See [Feature: Getting the size of a storage volume](https://gitlab.com/nunet/test-suite/-/blob/develop/stages/functional_tests/features/device-management-service/storage/basic_controller/Get_Size.feature)

For list of steps during execution, refer to the sequence diagram files:
* [mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/storage/basic_controller/sequences/getSize.sequence.mermaid)
* [svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/storage/basic_controller/sequences/rendered/getSize.sequence.svg)

# Custom configuration Parameters

Both `CreateVolume` and `LockVolume` allow for custom configuration of storage volumes via optional parameters. Below is the list of available parameters that can be used:

`WithPrivate()` - Passing this as an input parameter designates a given volume as private. It can be used both when creating or locking a volume.

`WithCID(cid string)` - This can be used as an input parameter to set the CID of a given volume during the lock volume operation.

# List of Data Types

`dms.storage.basic_controller.BasicVolumeController`: This struct manages implementation of `VolumeController` interface methods. See [basicVolumeController.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/storage/basic_controller/data/basicVolumeController.data.go) for reference data model.

Refer to package [readme](https://gitlab.com/nunet/device-management-service/-/blob/develop/storage/README.md#list-of-data-types) for other data types.




