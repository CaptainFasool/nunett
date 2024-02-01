## Introduction

This file explains the onboarding functionality of Device Management Service (DMS). This functionality is catered towards compute providers who wish provide their hardware resources to Nunet for running computational tasks as well as developers who are contributing to platform development.

### Stucture and organisation

Here is quick overview of the contents of this directory:

* [README](README.md): Current file which is aimed towards developers who wish to modify the onboarding functionality and build on top of it. 

* [handler](handler.go): This is main file where the code for onboarding functionality exists.

* [addresses](addresses.go): This file houses functions to generate Ethereum and Cardano wallet addresses along with its private key. 

* [addresses_test](addresses_test.go): This file houses functions to test the address generation functions defined in [addresses](addresses.go).

* [available_resources](available_resources.go): This file houses functions to get the total capacity of the machine being onboarded. 

* [init](init.go): This files initializes the loggers associated with onboarding package.

## Contributing

For guidelines of how to contribute, install and test the `device-management-service` component which contains `onboarding` package, please refer to package level documentation:

* Top level [../README.md](../README.md)
* Contribution guidelines [../CONTRIBUTING.md](../CONTRIBUTING.md);
* Code of conduct [../CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md)
* [Secure coding guidelines](https://gitlab.com/nunet/documentation/-/wikis/secure-coding-guidelines)

## Specifications and functionality

* The specification of package functionality is described as test case definitions, maintained in repository [test-suite/device-management-service/](https://gitlab.com/nunet/test-suite).
* The associated data models are are specified and maintained in repository [open-api/platform-data-model/device-management-service/onboarding](https://gitlab.com/nunet/open-api/platform-data-model/-/tree/develop/device-management-service/onboarding). 

Versioning and lifecycle of above mentioned specifications is aligned to the lifecycle and branching of the platform code (see [branching strategy](https://gitlab.com/nunet/documentation/-/wikis/GIT-Workflows#git-workflow-branching-strategy)):

* `develop` branches contain specifications of the functionality of current unstable branch of development at any given moment;
* `main` branches contain specifications of the current production version of the platform at any given moment in time;
* `proposed` branches are reserved only for test-suite and platform-data-model repositories which and contain new functionality and data model specifications, accepted for development, but not yet implemented and merged to any above branch.

Links to these two branches are provided with each interface endpoint description as applicable.

## Interface endpoints

### Onboard Compute Provider

**endpoint**: `/onboarding/onboard`<br/>
**methosd**: `HTTP POST`<br/>
**output**: `Machine Metadata`

This endpoint executes the onboarding process for a compute provider device. See table below for links to the onboarding specification and data models. 

| Spec type              | this branch     | proposed  |
---|---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/develop/stages/functional_tests/device-management-service/features/Onboard_Compute_Provider.feature))   | n.a. |
| Data (at rest)       | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/data/),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/data/rendered/)) | n.a. | 
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/onboardingProcess.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/rendered/onboardingProcess.sequence.svg)) | n.a. | 

> **_Note:_**  Indicated data models and structures should be understood as the result of the process. The onboarding process contains subprocesses which are reponsible for constructing the indicated data structures specifically.

DMS communicates to two external components during the onboarding process

* **Elasticsearch**: Currently used to log status updates. It is proposed that benchmarking data to be stored here

* **Logbin**: During onboarding, the device is only registered with this component. The main usage of Logbin is to monitor and record logs generated during execution of jobs.

### Get Device Info 

**endpoint**: `/onboarding/metadata`<br/>
**methods**: `HTTP GET`<br/>
**output**: `Machine Metadata`

This endpoint fetches the current metadata of the onboarded device.  

> **_Note:_** This endpoint is not being utilised at present. Instead this functionality is executed directly through a `ReadMetadataFile` method implementation in `device-management-service/utils/utils.go`

See table below for links to the specification and data models.
 
| Spec type              | this branch     | proposed  |
---|---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/develop/stages/functional_tests/device-management-service/features/Get_Device_Info.feature))   | n.a. | 
| Return payload       | entityDiagrams (TBD) | n.a. | 
| Request payload      | entityDiagrams (TBD) | n.a. | 
| Processes / Functions | sequenceDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/getMetadata.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/rendered/getMetadata.sequence.svg)) | n.a. | 

### Get Provisioned Capacity

**endpoint**: `/onboarding/provisioned`<br/>
**methods**: `HTTP GET`<br/>
**output**: `Provisioned Capacity`

This endpoint fetches the total capacity of the machine that is onboarded to Nunet.

See table below for links to the specification and data models.
 
| Spec type              | this branch     | proposed  |  
---|---|---|
| Features / test case specifications | Senarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/develop/stages/functional_tests/device-management-service/features/Get_Provisioned_Capacity.feature))   | n.a. | 
| Return payload       | entityDiagrams (TBD) | n.a. | 
| Request payload      | entityDiagrams (TBD) | n.a. | 
| Processes / Functions | sequenceDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/getProvisionedCapacity.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/rendered/getProvisionedCapacity.sequence.svg)) | n.a. | 

### Create Payment Address

**endpoint**: `/onboarding/address/new`<br/>
**methods**: `HTTP GET`<br/>
**output**: `Public-Private key pair & Mnemonic`

This endpoint creates a new blockchain payment address for the user.

See table below for links to the specification and data models.
 
| Spec type              | `develop`     | `proposed`  |   
---|---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/develop/stages/functional_tests/device-management-service/features/Create_Payment_Address.feature))   | n.a. | 
| Return payload       | entityDiagrams (TBD) | n.a. | 
| Request payload      | entityDiagrams (TBD) | n.a. | 
| Processes / Functions | sequenceDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/createPaymentAddress.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/rendered/createPaymentAddress.sequence.svg)) | n.a. | 

### Get Onboarding Status

**endpoint**: `/onboarding/status`<br/>
**method**: `HTTP GET`<br/>
**output**: `models.OnboardingStatus`

This endpoint returns onboarding status of the machine along with some metadata.

See table below for links to the specification and data models.
 
| Spec type              | this branch     | proposed  |
---|---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/develop/stages/functional_tests/device-management-service/features/Get_Onboarding_Status.feature))   | n.a. |
| Return payload       | entityDiagrams (TBD) | n.a. | 
| Request payload      | entityDiagrams (TBD) | n.a. | 
| Processes / Functions | sequenceDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/onboardingStatus.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/rendered/onboardingStatus.sequence.svg)) | n.a. |

### Change ResourceConfig

**endpoint**: `/onboarding/resource-config`<br/>
**methods**: `HTTP POST`<br/>
**output**: `Metadata`

This endpoint allows the user to change the configuration of the resources onboarded to Nunet.

See table below for links to the specification and data models.
 
| Spec type              | this branch     | proposed  | 
---|---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/develop/stages/functional_tests/device-management-service/features/Change_Resource_Config.feature))   | n.a. |
| Return payload       | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/messages/successfulResourceChange.message.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/messages/rendered/successfulResourceChange.message.svg))* | n.a. |
| Request payload      | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/messages/resourceChangeStart.message.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/messages/rendered/resourceChangeStart.message.svg))* | n.a. |
| Processes / Functions | sequenceDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/changeResource.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/rendered/changeResource.sequence.svg)) | n.a. | 


### Offboard

**endpoint**: `/onboarding/offboard`<br/>
**methods**: `HTTP DELETE`<br/>
**output**: `Success Message & Forced parameter`

This endpoint allows the user to remove the resources onboarded to Nunet. It provides flexibility by allowing a forced offboarding even in the presence of errors. The force parameter helps handle situations where it might be necessary to proceed with offboarding despite encountering issues.

See table below for links to the specification and data models.
 
| Spec type              | this branch     | proposed  | 
---|---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/develop/stages/functional_tests/device-management-service/features/Offboard.feature))   | n.a. | 
| Return payload (error)      | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/messages/offboardingError.message.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/messages/rendered/offboardingError.message.svg)) | n.a. |
| Return payload (success)      | entityDiagrams (TBD) | n.a. |
| Request payload      | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/messages/offboardingStart.message.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/messages/rendered/offboardingStart.message.svg)) | n.a. |
| Processes / Functions | sequenceDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/offboard.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/develop/device-management-service/onboarding/sequences/rendered/offboard.sequence.svg)) | n.a. |
