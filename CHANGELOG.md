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
# [0.4.75](#176)

### Added
- Respond with {"action": "job-submitted"} just before deployment request is handed over to compute provider
# [0.4.74](!135)

### Added
- Generate Events whenever there is an update to JobStatus field in service table

### Changed
- Send events on the same stream which is responsible for deployment response
- Update DeploymentRequestFlat table with JobStatus on service provider side
- Raise an error on frontend if a job is already running
- Close libp2p network stream when job is done either with success or failure
# [0.4.73](#175)

### Changed
- Increase timeout when connecting to Oracle
# [0.4.72](#173)

### Changed
- Bug fix: disable local address filter when not on server mode
## [0.4.71](!137)

### Changed
- No error on calcUsedResources when no services

## [0.4.70](#170)

### Changed
- Avoid error on unreplied ping
- Fix `dial to self attempted` bug
- ntx_payment event on claim by CP instead of on depReq by SP
- Avoid panic on deployment errors and send appropriate deployment response
- Avoid panic on gist update and dht update errors

### Added
- Oracle instances for channels other than team channel
## [0.4.69](#168)

### Changed
- Obtain gist token from a public endpoint
- Send depResp and close stream if unable to create gist
## [0.4.68](!129)

### Changed
- Raise HTTP exceptions for internal DMS error occurred on route handlers
## [0.4.67](#165)

### Changed
- Added ServiceCall event back after removed by mistake
## [0.4.66](#167)

### Changed
- Refactored statsdb device resource change calls
## [0.4.65](#164)

### Changed
- Avoid fetching peers at multiple places
- Increase relay limits and reservation resources
- Use only NuNet peers for relay and DHT update 

### Removed
- Removed the libp2p bootstrap nodes
## [0.4.64](#157)

### Added
- Rest endpoint to change amount of resource onboarded to NuNet
- Subcommand `resource-config`
- StatsDB event call on onboarded resource amount change
## [0.4.63](#161)

### Changed
- Make sure deviceinfo is set for outlier machines such as VMs (#161 bug fix)
## [0.4.62](#149)

### Changed
- Enabled autorelay and holepunching
- Upgraded libp2p and golang versions to 0.27 and 1.20
## [0.4.61](#109)

### Changed
- Re-establish connection to peers on network change
- Ping peers before sending Deployment requests
- Added functions to set up PubSub communication
## [0.4.60](#134)

### Added
- implemented ntx_payment event for sending transaction info to stats database

## [0.4.59](!114)

### Changed
- Delete entries in services table after response with oracle
## [0.4.58](#128)

### Changed
- Machine uuid for identification before onboarding and peerID attribute for opentelemetry
- Fix logger configuration for opentelemetry
- Remove unncessary env check for debug print and minor log format fix

## [0.4.57](!113)

### Changed
- Fix for base64 issue (cause for gist not being received by webapp)
- Improve log formatting
## [0.4.56](!112)

### Changed
- Debug mode DHT update interval env var and logging fix
## [0.4.55](!111)

### Changed
- Decode base64 message coming from libp2p stream
## [0.4.54](#153)

### Changed
- Set peerInfo struct to empty instead of removing peer from peerstore
## [0.4.53](#154)

### Changed
- Mark all records as deleted when forwarding deployment request to compute provider.
- Read first non-deleted record when reading temporary record.
## [0.4.52](#150)

### Changed
- Added List DHT peers endpoint
- Remove old peers from DHT
## [0.4.51](!106)

### Changed
- Set JobStatus to be finished with errors if container exited with non-zero exit status.
- Set JobStatus to be finished without errors if container exited with zero exit status.
- Return 102 status code when container is still running. DMS won't contact Oracle in such case.
## [0.4.50](!104)

### Changed
- add dht/peers route and debug print on deployment request
- bug fixes and debug prints with debug env var
## [0.4.49](!102)

### Changed
- filter machines for cpu only ml jobs
- fence index error
## [0.4.48](#143)

### Changed
- Fix for problem with CPU-only job deployment
## [0.4.47](#142)

### Changed
- Save LogURL and other service related data on service run
- Fetch LogURL and other service related data on request reward
## [0.4.46](!96)

### Changed
- Correct the wrong formatting of deployment response.
- Minor refactoring
## [0.4.45](#144)

### Changed
- Quick install script to query for blockchain when creating wallet
- Fix instruction on readme to avoid running in a subshell
- Use NuNet bootstrap servers and default libp2p host options 
## [0.4.44](!94)

### Changed
- Fix for the problem with non-flat model db migration
## [0.4.43](#123)

### Changed
- DMS integration with Oracle to get blockchain data
  
## [0.4.42](#139)

### Changed
- Replaced GPU detection library 
  
## [0.4.41](#125)

### Added
- The logic to handel the choice of CPU and GPU

## [0.4.40](#137)

### Added
- Reduced DHT update interval

### Removed
- Calls to StatsDB in GPU deployment. Refer to #138

## [0.4.39](#124)

### Added
- Included pre-commit hook and a dev-setup.sh.
### Changed
- Fix for CORS error.
## [0.4.38](#133)

### Added
- Use public key from DB instead of attempting to extract from ID
## [0.4.37](#105)

### Added
- NuNet Hosted Bootstrap Servers
## [0.4.36](#135)

### Added
- Server mode to disable local network scan on datacenters
## [0.4.35](#132)

### Added
- Detect GPU while onboarding and update GPU flag on DHT

## [0.4.34](#115)

### Added
- Instrumentation
## [0.4.33](#119)

### Removed
- Removed all NuNet Adapter usage and installation
## [0.4.32](#129)

### Added
- Sending DHT updates periodically
## [0.4.31](#107)

### Added
- Deployment request with libp2p messaging
- Simple chat between peers using CLI
## [0.4.30](#106)

### Added
- Implemented DHT on libp2p with similar schema to the nunet adapter
- Helper functions to query the libp2p DHT

## [0.4.29](nunet/ml-on-gpu/ml-on-gpu-service#14)
### Changed
- Updated nunet info command to detect GPUs
- Updated nunet onboard-ml to detect ML images after the command is run
### Added
- New nunet available --gpu-status command to show GPUs metrics in real-time
- New nunet available --cuda-tensor command to check availability of CUDA and Tensor Cores

## [0.4.28](#104)

### Changed
- Improvement on discoverability
- Increase deadline on getSelfNodeID 
- Remove unnecessary code
### Added
- Missing Dependencies on Deb Package

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

