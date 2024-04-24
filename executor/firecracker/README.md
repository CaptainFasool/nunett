# Introduction
This sub-package contains functionality including drivers and api for the Firecracker executor.

# Stucture and organisation

_proposed 2024-04-22; by @0xPravar; @dawit.abate_

Here is quick overview of the contents of this pacakge:

* [README](README.md): Current file which is aimed towards developers who wish to use and modify the Firecracker functionality. 

* [client](client.go): This file provides a high level wrapper around the [Firecracker](github.com/firecracker-microvm/firecracker-go-sdk) library.

* [executor](executor.go): This is the main implementation of the executor interface for Firecracker. It is the entry point of the sub-package. It is intended to be used as a singleton.

* [handler](handler.go): This file contains a handler implementation to manage the lifecycle of a single job.

* [init](init.go): This file is responsible for initialization of the package. Currently it only initializes a logger to be used through out the sub-package.

* [types](types.go): This file contains Models that are specifically related to the Firecracker executor. Mainly it contains the engine spec model that describes a Firecracker job.

# Contributing

_proposed 2024-04-22; by @0xPravar; @dawit.abate_

For guidelines of how to contribute, install and test the `device-management-service` component which contains `executor` package, please refer to package level documentation:

* Executor package level [../README.md](../README.md)
* DMS component level [../../README.md](../../README.md)
* Contribution guidelines [../../CONTRIBUTING.md](../../CONTRIBUTING.md)
* Code of conduct [../../CODE_OF_CONDUCT.md](../../CODE_OF_CONDUCT.md)
* [Secure coding guidelines](https://gitlab.com/nunet/documentation/-/wikis/secure-coding-guidelines)

# Specifications overview

_proposed 2024-04-22; by @0xPravar; @dawit.abate_

Refer to the [specifications overview](../README.md#specifications-overview) in the package readme.

# Functions

### NewExecutor

_proposed 2024-04-22; by @0xPravar; @dawit.abate_

* signature: `NewExecutor(_ context.Context, id string) -> (dms.executor.firecracker.Executor, error)` <br/>
* input #1: `Go context` <br/>
* input #2: identifier of the executor <br/>
* output (sucess): Executor instance of type `dms.executor.firecracker.Executor` <br/>
* output (error): error

`NewExecutor` function initializes a new Executor instance for Firecracker VMs. 

It is expected that `NewExecutor` would be called prior to calling any other executor functions. The Executor instance returned would then be used to call other functions like `Start`, `Stop` etc.

### IsInstalled

_proposed 2024-04-22; by @0xPravar; @dawit.abate_

For function signature refer to the package [readme](../README.md#isinstalled)

`IsInstalled` checks if the Firecracker is installed on the host. It returns `true` if Firecracker is installed and accessible, `false` otherwise. 

See [Feature: Firecracker Installation Check](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/executor/firecracker/Is_Installed.feature)

### Start

_proposed 2024-04-23; by @0xPravar; @dawit.abate_

For function signature refer to the package [readme](../README.md#start) 

`Start` function begins the execution of a request by starting a Firecracker VM. It creates the VM based on the configuration parameters provided in the execution request. It returns an error message if
* execution is already started
* execution is already finished
* there is failure is creation of a new VM

See [Feature: Start Firecracker VM](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/executor/firecracker/Start.feature)

### Wait

_proposed 2024-04-23; by @0xPravar; @dawit.abate_

For function signature refer to the package [readme](../README.md#wait)

`Wait` initiates a wait for the completion of a specific execution using its `executionID`. The function returns two channels: one for the result and another for any potential error. 

If the `executionID` is not found, an error is immediately sent to the error channel.

Otherwise, an internal goroutine is spawned to handle the asynchronous waiting. The entity calling should use the two returned channels to wait for the result of the execution or an error. If there is a cancellation request (context is done) before completion, an error is relayed to the error channel. When the execution is finished, both the channels are closed.

See [Feature: Wait for a execution](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/executor/firecracker/Wait.feature)

### Cancel

_proposed 2024-04-23; by @0xPravar; @dawit.abate_

For function signature refer to the package [readme](../README.md#cancel)

`Cancel` tries to terminate an ongoing execution identified by its `executionID`. It returns an error if the execution does not exist.

See [Feature: Cancel execution](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/executor/firecracker/Cancel.feature)

### Run

_proposed 2024-04-23; by @0xPravar; @dawit.abate_

For function signature refer to the package [readme](../README.md#run)

`Run` initiates and waits for the completion of an execution in one call. This method serves as a higher-level convenience function that internally calls `Start` and `Wait` methods. It returns the result of the execution as `dms.executor.ExecutionResult` type. 

It returns an error in case of:
* failure in starting the VM
* failure in waiting 
* context is cancelled

See [Feature: Run Execution](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/executor/firecracker/Run.feature)

### Cleanup

_proposed 2024-04-23; by @0xPravar; @dawit.abate_

* signature: `Cleanup(ctx context.Context) -> error` <br/>
* input: `Go context` <br/>
* output (sucess): None <br/>
* output (error): error

`Cleanup` removes all firecracker resources associated with the executor. This includes stopping and removing all running VMs and deleting their socket paths. It returns an error it it is unable to remove the containers.

See [Feature: Cleanup firecracker Resources](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/executor/firecracker/Cleanup.feature)

# List of Data Types

_proposed 2024-04-23; by @0xPravar; @dawit.abate_

`dms.executor.firecracker.Executor`: This is the instance of the executor created by `NewExecutor` function. It contains the firecracker client and other resources required to execute requests. See [executor.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/executor/firecracker/data/executor.data.go) for reference data model.

`dms.executor.firecracker.executionHandler`: This contains necessary information to manage the execution of a firecracker VM. See [executionHandler.data.go](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/executor/firecracker/data/executionHandler.data.go) for reference data model.

Refer to package [readme](../README.md#list-of-data-types) for other data types.