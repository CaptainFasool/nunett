## Introduction

This file explains the api functionality of Device Management Service (DMS). DMS exposes various endpoints through which its different functionalities can be accessed. 

## Stucture and organisation

Here is quick overview of the contents of this directory:

* [README](README.md): Current file which is aimed towards developers who wish to use and modify the api functionality. 

* [api](api.go): This file contains router setup using Gin framework. It also applies Cross-Origin Resource Sharing (CORS) middleware and OpenTelemetry middleware for tracing. Further it lists down the endpoint URLs and the associated handler functions.

* [debug](debug.go): This file contains endpoints which are only available when `DEBUG` mode is enabled.

* [device](device.go): This file contains endpoints to retrieve and modify the device status.

* [onboarding](onboarding.go): This file contains endpoints related to the onboarding functionality catered towards compute providers.

* [peers](peers.go): This file contains various endpoints related to the p2p functionality of DMS. 

* [run](run.go): This file contains various endpoints related to the deployment and execution of jobs.

* [telemetry](telemetry.go): This file contains the endpoint to calculate available free resources in a machine.

* [transactions](transactions.go): This file contains the endpoints related to blockchain transactions.

* [vm](vm.go): This file contains the endpoints related to starting a [firecracker VM](https://firecracker-microvm.github.io/) with custom or default configuration.

All of these files have a counterpart named as `*_test.go` which contains the unit tests for the corresponding endpoints.

## Contributing

For guidelines of how to contribute, install and test the `device-management-service` component which contains `api` package, please refer to package level documentation:

* Top level [../README.md](../README.md)
* Contribution guidelines [../CONTRIBUTING.md](../CONTRIBUTING.md);
* Code of conduct [../CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md)
* [Secure coding guidelines](https://gitlab.com/nunet/documentation/-/wikis/secure-coding-guidelines)

## Specifications overview

* The specification of package functionality is described as test case definitions, maintained in repository [test-suite](https://gitlab.com/nunet/test-suite).
* The associated data models are specified and maintained in repository [open-api/platform-data-model/device-management-service/](https://gitlab.com/nunet/open-api/platform-data-model/-/tree/develop/device-management-service/). 

Versioning and lifecycle of above mentioned specifications is aligned to the lifecycle and branching of the platform code (see [branching strategy](https://gitlab.com/nunet/documentation/-/wikis/GIT-Workflows#git-workflow-branching-strategy)):

* `develop` branches contain specifications of the functionality of current unstable branch of development at any given moment;
* `main` branches contain specifications of the current production version of the platform at any given moment in time;
* `proposed` branches contain new functionality and data model specifications, accepted for development, but not yet implemented.

The procedure to update the specifications is described in [Specification And Documentation Procedure](https://gitlab.com/nunet/team-processes-and-guidelines/-/blob/main/Specification_And_Documentation_Procedure.md?ref_type=heads).

## API functionality

The following sections describe the different functionality of the DMS covered in the `api` package.

### device

#### Device Status

**endpoint**: `/device/status`<br/>
**method**: `HTTP GET`<br/>
**output**: `Device Status`

The endpoint retrieves the current status of the machine (online / offline).

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/device/Device_Status.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](),[.svg]()) | 

#### Change Device Status

### onboarding

### peers

### run

### telemetry

### transactions

### vm

### debug


