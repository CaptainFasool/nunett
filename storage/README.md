# Introduction
_proposed 2024-04-17; by @0xPravar; @joao.castro5_

The storage package is responsible for disk storage management on each DMS (Device Management Service) for data related to DMS and jobs deployed by DMS. It primarily handles storage access to remote storage providers such as [AWS S3](https://aws.amazon.com/s3/), [IPFS](https://ipfs.tech/) etc. It also handles the control of storage volumes. 

# Interfaces

### StorageProvider

_proposed 2024-04-17; by @0xPravar; @joao.castro5_

The `StorageProvider` interface handles the input and output operations of files with remote storage providers such as AWS S3 and IPFS. Basically it provides methods to upload or download data and also to check the size of a data source.

Its functionality is coupled with local mounted volumes, meaning that implementations will rely on mounted files to upload data and downloading data will result in a local mounted volume.

*Notes:* 
* If needed, the availability-checking of a storage provider should be handled druing instantiation of the implementation. 

* Any necessary authentication data should be provided within the `dms.executor.SpecConfig` parameters

* The interface has been designed for file based transfer of data. It is not built with the idea of supporting streaming of data and non-file storage operations (e.g.: some databases). Assessing the feasiblity of such requirement if needed should be done while implementation. 

See [storageProvider.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/storage/data/storageProvider.data.go) for the current reference model.

### VolumeController

_proposed 2024-04-17; by @0xPravar; @joao.castro5_

The `VolumeController` interface manages operations related to storage volumes which are data mounted to files/directories. Its methods allow for creation or deletion of volumes. It also allows to fetch the list of volumes that already exist.

See [volumeController.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/storage/data/volumeController.data.go) for the current reference model.

# Types

_proposed 2024-04-17; by @0xPravar; @joao.castro5_

`dms.storage.StorageVolume`: This struct contains parameters related to a storage volume such as path, volume type etc. See [storageVolume.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/storage/data/storageVolume1.data.go) for reference data model. (_Note: [storage volume](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/storage/data/storageVolume.data.go) is also defined in `executor` package. This needs to be resolved_)

`dms.executor.SpecConfig`: This allows arbitrary configuration/parameters as needed during implementation of a specific storage provider. The parameters include authentication related data (if applicable). See [specConfig.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/executor/data/specConfig.data.go) for reference data model.

`dms.storage.error`: This contains error details returned by storage operations. See [error.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/storage/data/error.data.go) for reference data model.

# Functions

### Upload

_proposed 2024-04-17; by @0xPravar; @joao.castro5_

* signature: `Upload(vol dms.storage.StorageVolume, target dms.executor.SpecConfig) -> (dms.executor.SpecConfig, error)` <br/>
* input #1: storage volume from which data will be uploaded of type `dms.storage.StorageVolume` <br/>
* input #2: configuration parameters of specified storage provider of type `dms.executor.SpecConfig` <br/>
* output (sucess): parameters related to storage provider like upload details/metadata etc of type `dms.executor.SpecConfig` <br/>
* output (error): error of type `dms.storage.error`

`Upload` function uploads data from the storage volume provided as input to a given remote storage provider. The configuration of the storage provider is also provided as input to the function.

It is expected that the return value will vary from provider to provider (and in some cases it may be NIL) and hence it is left to be detailed during implementation. 

It will return an error if the operation fails. 

See [Feature: Upload data](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/storage/Upload.feature) for the different scenarios.

### Download

_proposed 2024-04-17; by @0xPravar; @joao.castro5_

* signature: `Download(source dms.executor.SpecConfig, outputPath string) -> (dms.storage.StorageVolume, error)` <br/>
* input #1: configuration parameters of specified storage provider of type `dms.executor.SpecConfig` <br/>
* input #2: output path where downloaded data should be stored <br/>
* output (sucess): storage volume which has downloaded data of type `dms.storage.StorageVolume` <br/>
* output (error): error of type `dms.storage.error`

`Download` function downloads data from a given source, mounting it to a certain local path. The input configuration received will vary from provider to provider and hence it is left to be detailed during implementation.

It will return an error if the operation fails. Note that this can also happen if the user running DMS does not have access permission to the given path.

See [Feature: Download data](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/storage/Download.feature) for the different scenarios.

### Size

_proposed 2024-04-17; by @0xPravar; @joao.castro5_

* signature: `Size(source dms.executor.SpecConfig) -> (MiB, error)` <br/>
* input: configuration parameters of specified storage provider of type `dms.executor.SpecConfig` <br/>
* output (sucess): size of the storage in Megabytes of type `uint64` <br/>
* output (error): error of type `dms.storage.error`

`Size` function returns the size of a given storage provider provided as input. It will return an error if the operation fails.

See [Feature: Get size of the storage](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/storage/Size.feature) for the different scenarios.

Note that this method may also be useful to check if a given source is available.

### CreateVolume

_proposed 2024-04-17; by @0xPravar; @joao.castro5_

* signature: `CreateVolume(path string, volType string) -> (dms.storage.StorageVolume, error)` <br/>
* input #1: path to the location where volume should be created <br/>
* input #2: type of volume to be created <br/>
* output (sucess): storage volume of type `dms.storage.StorageVolume` <br/>
* output (error): error of type `dms.storage.error`

`CreateVolume` function creates a new volume at the provided path. The type of volume to be created also needs to specified in the input. It returns a storage volume struct containing details of the newly created volume. 

It will return an error if the operation fails. Note that this can also happen if the user running DMS does not have access permission to create volume at the given path.

See [Feature: Create Volume](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/storage/Create_Volume.feature) for the different scenarios.

### DeleteVolume

_proposed 2024-04-17; by @0xPravar; @joao.castro5_

* signature: `DeleteVolume(in dms.storage.StorageVolume) -> error` <br/>
* input: storage volume to be deleted of type `dms.storage.StorageVolume` <br/>
* output (sucess): None <br/>
* output (error): error of type `dms.storage.error`

`DeleteVolume` function deletes the specified storage volume. It will return an error if the operation fails. Note that this can also happen if the user running DMS does not have the requisite access permissions.

See [Feature: Delete Volume](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/storage/Delete_Volume.feature) for the different scenarios.

### ListVolumes

_proposed 2024-04-17; by @0xPravar; @joao.castro5_

* signature: `ListVolumes() -> ([]dms.storage.StorageVolume, error)` <br/>
* input: None <br/>
* output (sucess): List of existing storage volumes of type `dms.storage.StorageVolume` <br/>
* output (error): error of type `dms.storage.error`

`ListVolumes` function fetches the list of existing storage volumes. It will return an error if the operation fails or if the user running DMS does not have the requisite access permissions.

See [Feature: List existing storage volumes](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/storage/List_Volume.feature) for the different scenarios.