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

## [0.4.7](#047)

### Added
- Send docker logs for running container to GitHub's Gist

## [0.4.6](#046)

### Added
- Added stats_db grpc calls.

## [0.4.5](#045)

### Added
- Added a grpc-client to access cardano-cli
- Added a function interface to run cardano-cli commands
- Removed unused functions

## [0.4.4](#044)

### Added
- Added command for installing NVIDIA GPU driver and Container Runtime in nunet CLI.
- Added command for pre-downloading ML docker images in nunet CLI. 

## [0.4.3](#043)

### Fixed
- Fix 500 response on `/vm/*` endpoints.

### Removed
- Remove internal endpoints.

### Changed
- Refactor `vm` module for speed and readability.
- the websocat installation to include $archdir

## [0.4.2](#0412)

### Added
- Trigger docker deployment 
- the shell command which takes node id 

### Removed
- Don't remove image after GPU load is run.

### Fixed
- Fix one logical error where DMS on compute provider sent error ack for expired message. 

