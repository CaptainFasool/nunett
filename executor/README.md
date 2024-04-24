# Introduction
The executor package is responsible for executing the jobs received by the device management service (DMS). It provides an unified interface to run various executors such as docker, firecracker etc

# Stucture and organisation

_proposed 2024-04-17; by @0xPravar; @dawit.abate_

Here is quick overview of the contents of this pacakge:

* [README](README.md): Current file which is aimed towards developers who wish to use and modify the executor functionality. 

* [init](init.go): This file initializes a logger instance for the executor package.

* [types](types.go): This file contains the interfaces that other packages in the DMS call to utilise functionality offered by the executor package.

* [docker](docker): This folder contains the implementation of docker executor.
 
# Contributing

_proposed 2024-04-17; by @0xPravar; @dawit.abate_

For guidelines of how to contribute, install and test the `device-management-service` component which contains `executor` package, please refer to package level documentation:

* DMS component level [../README.md](../README.md)
* Contribution guidelines [../CONTRIBUTING.md](../CONTRIBUTING.md)
* Code of conduct [../CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md)
* [Secure coding guidelines](https://gitlab.com/nunet/documentation/-/wikis/secure-coding-guidelines)

# Specifications overview

_proposed 2024-04-17; by @0xPravar; @dawit.abate_

* The specification of package functionality is described as test case definitions, maintained in repository [test-suite](https://gitlab.com/nunet/test-suite).
* The associated data models are specified and maintained in repository [open-api/platform-data-model/device-management-service/](https://gitlab.com/nunet/open-api/platform-data-model/-/tree/develop/device-management-service/). 

Versioning and lifecycle of above mentioned specifications is aligned to the lifecycle and branching of the platform code (see [branching strategy](https://gitlab.com/nunet/documentation/-/wikis/GIT-Workflows#git-workflow-branching-strategy)):

* `develop` branches contain specifications of the functionality of current unstable branch of development at any given moment;
* `main` branches contain specifications of the current production version of the platform at any given moment in time;
* `proposed` branches contain new functionality and data model specifications, accepted for development, but not yet implemented.

The procedure to update the specifications is described in [Specification And Documentation Procedure](https://gitlab.com/nunet/team-processes-and-guidelines/-/blob/main/Specification_And_Documentation_Procedure.md?ref_type=heads).

# Executor Functionality

_proposed 2024-04-17; by @0xPravar; @dawit.abate_

The main functionality offered by the `executor` package is defined via the `Executor` interface. It has following methods:

### IsInstalled

* signature: `IsInstalled(ctx context.Context) -> bool` <br/>
* input: `Go context` <br/>
* output: `bool` 

`IsInstalled` checks if the executor is installed and available for use. It takes the Go `context` object as input and returns a boolean indicating if the executor is installed or not.

### Start

* signature: `Start(ctx context.Context, request dms.executor.ExecutionRequest) -> error` <br/>
* input #1: `Go context` <br/>
* input #2: `dms.executor.ExecutionRequest` <br/>
* output: `error` 

`Start` function takes a Go `context` object and a `dms.executor.ExecutionRequest` type as input. It returns an error if the execution already exists and is in a started or terminal state. Implementations may also return other errors based on resource limitations or internal faults.

### Run

* signature: `Run(ctx context.Context, request dms.executor.ExecutionRequest) -> (dms.executor.ExecutionResult, error)` <br/>
* input #1: `Go context` <br/>
* input #2: `dms.executor.ExecutionRequest` <br/>
* output (success): `dms.executor.ExecutionResult` <br/>
* output (error): `error`

`Run` initiates and waits for the completion of an execution for the given Execution Request. It returns a `dms.executor.ExecutionResult` and an error if any part of the operation fails. Specifically, it will return an error if the execution already exists and is in a started or terminal state.

### Wait

* signature: `Wait(ctx context.Context, executionID string) -> (<-chan dms.executor.ExecutionResult, <-chan error)` <br/>
* input #1: `Go context` <br/>
* input #2: `dms.executor.ExecutionRequest.ExecutionID` <br/>
* output #1: Channel that returns `dms.executor.ExecutionResult` <br/>
* output #2: Channel that returns `error`

`Wait` monitors the completion of an execution identified by its `executionID`. It returns two channels:
1. A channel that emits the execution result once the task is complete;
2. An error channel that relays any issues encountered, such as when the execution is non-existent or has already concluded.

### Cancel

* signature: `Cancel(ctx context.Context, executionID string) -> error` <br/>
* input #1: `Go context` <br/>
* input #2: `dms.executor.ExecutionRequest.ExecutionID` <br/>
* output: `error`

`Cancel` attempts to terminate an ongoing execution identified by its `executionID`. It returns an error if the execution does not exist or is already in a terminal state.

### GetLogStream

* signature: `GetLogStream(ctx context.Context, request dms.executor.LogStreamRequest, executionID string) -> (io.ReadCloser, error)` <br/>
* input #1: `Go context` <br/>
* input #2: `dms.executor.LogStreamRequest` <br/>
* input #3: `dms.executor.ExecutionRequest.ExecutionID` <br/>
* output #1: `io.ReadCloser` <br/>
* output #2: `error`

`GetLogStream` provides a stream of output for an ongoing or completed execution identified by its `executionID`. There are two flags that can be used to modify the functionality:
* The `Tail` flag indicates whether to exclude historical data or not.
* The `follow` flag indicates whether the stream should continue to send data as it is produced.

It returns an `io.ReadCloser` object to read the output stream and an error if the operation fails. Specifically, it will return an error if the execution does not exist.

## List of Data Types

_proposed 2024-04-17; by @0xPravar; @dawit.abate_

`dms.executor.ExecutionRequest`: This is the input that `executor` receives to initiate a job execution. See [executionRequest.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/executor/data/executionRequest.data.go) for reference data model. (_Note: This needs to be aligned with the [dms.orchestrator.invocation](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/data/invocation.payload.go) data model_)

`dms.executor.ExecutionResult`: This contains the result of the job execution. See [executionResult.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/executor/data/executionResult.data.go) for reference data model. (_Note: This needs to be aligned with the [dms.jobs.result](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/jobs/data/result.payload.go) data model_)

`dms.executor.LogStreamRequest`: This contains input parameters sent to the `executor` to get job execution logs. See [logStreamRequest.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/executor/data/logStreamRequest.data.go) for reference data model.  

`dms.models.SpecConfig`: This allows arbitrary configuration/parameters as needed during implementation of specific executor. See [specConfig.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/models/data/specConfig.data.go) for reference data model.

`dms.executor.ExecutionResources`: This contains resources to be used for execution. See [executionResources.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/executor/data/executionResources.data.go) for reference data model. (_Note: This needs to be aligned with the resource specification used by orchestrator_)

`dms.storage.StorageVolume`: This contains parameters of storage volume used during execution. See [storageVolume.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/storage/data/storageVolume.data.go) for reference data model.
