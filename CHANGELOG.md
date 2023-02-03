<!-- New changes go on top. But below these comments. -->

<!-- We can track changes from Git logs, so why we needs this file? 

Guiding Principles
- Changelogs are for humans, not machines.
- There should be an entry for every single version.
- The same types of changes should be grouped.
- Versions and sections should be linkable.
- The latest version comes first.
- The release date of each version is displayed.
- Mention whether you follow Semantic Versioning.

Types of changes
- `Added` for new features.
- `Changed` for changes in existing functionality.
- `Deprecated` for soon-to-be removed features.
- `Removed` for now removed features.
- `Fixed` for any bug fixes.
- `Security` in case of vulnerabilities.

-->
## [0.4.27](#103)

### Added
- Libp2p node 
-  `nunet peer list` prints peers found by the libp2p node 

## [0.4.26](#92)
### Changed

Send the deployment response container log gist URL to GPU ML user.

## [0.4.25](#111)

### Added
- Unit tests for CLI

## [0.4.24](#85)

### Added
- Unit tests for /onboarding REST endpoints

## [0.4.23](#90)

### Changed
- Separated StatsDB GRPC address for different onboarding channels

## [0.4.22](#81)

### Added
- Enable wallet creation for the Cardano blockchain

## [0.4.21](#93)

### Changed
- Enable GPU onboarding for WSL users

## [0.4.20](#95)

### Changed
- Refactor logger

## [0.4.19](#75)

### Changed
- Remove random amount of wait in `Onboard` handler.
- Move telemetry code to `InstallRunAdapter` for faster request-response cycle.

## [0.4.18](#89)

### Fixed
- Fixed VM network config

## [0.4.17](documentation#19)

### Added
- CLI command to collect log and return path

## [0.4.16](#80)

### Added
- CLI command to collect log and return path

## [0.4.15](#63)

### Added
- Implementation of deployement request status information propagation to statsdb for ML job deployment.

## [0.4.14](#78)

### Added
- New gRPC service for retreiving Master Community's public key

## [0.4.13](#60)

### Added
- New grpc service for incoming messages and message receiver 
- Preliminary deployment request handler for cardano and gpu usecases

## [0.4.12](#69)

### Added
- shell command for websocat 
- websocat installation

## [0.4.11](#49)

### Added
- Passing fields required for Remote Shell authorization on VMs.

## [0.4.10](!41)

### Added
- Ping-pong websocket implementation. This connection will be used to send commands and receive output from remote shell.

## [0.4.9](#46)

### Added
- Calculate telemetry of docker containers and update DHT.
- Missing dependency `bc` in DEBIAN/control

### Changed 
- Moved DHT update grpc call to run in a separate thread.
- Image used in ci for building in order to support at least glibc-2.27

## [0.4.8](#73)

### Added
- Handle onboarding channels for separate networks corresponding to https://gitlab.com/nunet/nunet-adapter/-/issues/108
- Workaround for #74

## [0.4.7](#64)

### Added
- Send docker logs for running container to GitHub's Gist

## [0.4.6](#63)

### Added
- Added stats_db grpc calls.

## [0.4.5](#67)

### Added
- Added a grpc-client to access cardano-cli
- Added a function interface to run cardano-cli commands
- Removed unused functions

## [0.4.4](#61)

### Added
- Added command for installing NVIDIA GPU driver and Container Runtime in nunet CLI.
- Added command for pre-downloading ML docker images in nunet CLI. 

## [0.4.3](#66)

### Fixed
- Fix 500 response on `/vm/*` endpoints.

### Removed
- Remove internal endpoints.

### Changed
- Refactor `vm` module for speed and readability.
- the websocat installation to include $archdir

## [0.4.2](#57)

### Added
- Trigger docker deployment 
- the shell command which takes node id 

### Removed
- Don't remove image after GPU load is run.

### Fixed
- Fix one logical error where DMS on compute provider sent error ack for expired message. 

