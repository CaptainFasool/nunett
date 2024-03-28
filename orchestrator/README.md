## Introduction
This package takes care of job scheduling and management (manages jobs on other DMSs)

## Job Orchestration
The lifecyle of a job on Nunet platform consists of various operations from job posting to settlement of the contract. A key distinction to note is the option of two types of orchestration mechanisms: `push` and `pull`. Broadly speaking `pull` orchestration works on the premise that resource providers bid for jobs available in the network. 

Whereas `push` orchestration develops on the idea that users choose from the available providers and their resources. However, given the decentralised and open nature of the platform, it may be required to engage the providers to get their current (latest) state and preferences. This leads to an overlap with the `pull` orchestration approach.

The default setting is to use `pull` based orchestration. However, the user can choose to use `push` based orchestration to suit their needs.

The details of all the operations involved in the job orchestration are described in the following sections.

### 1. Job Posting

* proposed 2024-03-21; by: @kabir.kbr; @janaina.senna; @0xPravar; *

This is based on preliminary design, please refer to [research/blog/jobposting](https://nunet.gitlab.io/research/blog/posts/job-orchestration-details/#1-job-posting).

The first step is when a user posts a request to run a computing job. This should define various job requirements and preferences.

**endpoint**: `/orchestrator/postJob`<br/>
**method**: `HTTP POST`<br/>
**output**: `None`

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/orchestrator/Job_Posting.feature))   |
| Request payload       | [jobDescription](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/jobs/data/jobDescription.payload.go)|
| Return payload       | None |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/sequences/jobPosting.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/sequences/rendered/jobPosting.sequence.svg)) | 

#### List of relevant functions

* `dms.orchestrator.processJob()` - This function will validate the job received, add metadata (if needed) and save the job to the local database.

#### List of relevant data types
* `dms.jobs.jobDescription` - This contains the job details and desired capability needed to execute the job.


### 2. Search and Match

* proposed 2024-03-27; by: @kabir.kbr; @janaina.senna; @0xPravar; *

Once the DMS received a job posting, it will look to find nodes that can service the request. This is done by matching the job requirements with the available resources.

### Configuration
`dms.dms.config.defaultOrchestrationType`: The default setting of the network is use `pull` search and match operation. This value is stored in `defaultOrchestrationType` parameter saved in the [config](https://gitlab.com/nunet/device-management-service/-/tree/proposed/dms/config) folder under `dms` package. The user can change this value to `push` if needed via CLI (Command Line Interface) or API. This functionality is covered in the [dms](https://gitlab.com/nunet/device-management-service/-/tree/proposed/dms) folder.

`dms.dms.config.defaultSearchTimeout`: Each DMS will have default timeout value for the search operation. This value is stored in `defaultSearchTimeout` parameter saved in the [config](https://gitlab.com/nunet/device-management-service/-/tree/proposed/dms/config) folder under `dms` package. This can be overridden by the owner of the DMS via CLI (Command Line Interface) or API. This functionality is covered in the [dms](https://gitlab.com/nunet/device-management-service/-/tree/proposed/dms) folder.

### Pull Based
The first step is that service provider DMS requests bids from the compute providers in the network. DMS on compute provider compares the capability of the available resources against soft and hard constraints specified in the job requirements. If all the requirements are met, it then decides whether to submit a bid. The final outcome is that service provider DMS has a list of eligible compute providers with their bids.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/orchestrator/Pull_Search_And_Match.feature))   |
| Request payload       | [BidRequest](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/data/bidRequest.payload.go)|
| Data at rest (CP DMS)      | [AvailableCapability](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/dms/data/availableCapability.payload.go) |
| Data at rest (CP DMS)      | [CapabilityComparison](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/data/capabilityComparison.payload.go) |
| Return payload       | [Bid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/data/bid.payload.go) |
| Data at rest (SP DMS)       | [EligibleComputeProvidersIndex](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/data/computeProviderIndex.payload.go) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/sequences/pullSearchAndMatch.sequence.mermaid),[.svg]()) |

The second step is to shortlist the preferred compute provider peer based on some selection criteria. Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/orchestrator/Select_Preferred_Node.feature))   |
| Request payload       | [EligibleComputeProvidersIndex](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/data/computeProviderIndex.payload.go) |
| Return payload       | [EligibleComputeProviderData](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/data/computeProviderIndex.payload.go) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/sequences/selectPreferredNode.sequence.mermaid),[.svg]()) |

**List of relevant functions**:<br/>
`dms.network.publishJob()` - This function will take the `dms.orchestrator.bidRequest` as input and propogate it to the compute providers in the network.

`dms.dms.availableResources()` - This function will return `dms.dms.availableCapability` which is the available resources/capability of the machine to perform a job.

`dms.orchestrator.compare()` - This function takes `dms.jobs.jobDescription.requiredCapability` and `dms.dms.availableCapability` as input and returns `dms.orchestrator.capabilityComparison`.

`dms.orchestrator.accept()` - This function takes `dms.orchestrator.capabilityComparison` and `dms.dms.config.capabilityComparator` as inputs and returns a `bool` which indicates whether the job can be accepted or not.

`dms.orchestrator.createBid()` - This function returns `dms.orchestrator.bid` which is to be sent to service provider DMS.

`dms.orchestrator.registerComputeBid()` - This function takes `dms.orchestrator.bid` as inputs and saves it to a table in the local database as `dms.orchestrator.computeProvidersIndex`.

**List of relevant data types**:<br/>
`dms.orchestrator.bidRequest` - This is sent to the compute providers based on which they can submit a bid.

`dms.orchestrator.bid` - This is the bid that is submitted by the compute provider.

`dms.dms.availableCapability` - This contains the available resources/capability of the machine to perform a job.

`dms.orchestrator.capabilityComparison` - This contains the comparison of the job requirements with the available resources.

`dms.orchestrator.computeProvidersIndex` - This contains the data of compute providers whose bids have been received. 

### 3. Job Request

* proposed 2024-03-27; by: @kabir.kbr; @janaina.senna; @0xPravar; *

DMS on service provider side checks whether the resources are locked by the preferred compute provider peer.

In case the shortlisted compute provider has not locked the resources while submitting the bid, the job request workflow is executed. This requires the compute provider DMS to lock the necessary resources required for the job and re-submit the bid. Note that at this stage compute provider can still decline the job request.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/proposed/stages/functional_tests/features/device-management-service/orchestrator/Job_Request.feature))   |
| Request payload       | [BidRequest](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/data/bidRequest.payload.go) |
| Return payload - request acceptance      | [Bid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/data/bid.payload.go) |
| Return payload - request denied      | [DeclineMessage](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/data/declineJobRequest.payload.go) |
| Return payload - timeout      | [TimeoutResponse](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/data/timeoutJobRequest.payload.go) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/sequences/jobRequest.sequence.mermaid),[.svg]()) |

**List of relevant functions**:<br/>
`dms.orchestrator.checkResourceLock()` - This function takes a value from `dms.orchestrator.computeProvidersIndex` as input and checks if the resources required for the job are locked by the compute provider. It returns a `bool` value as output.

`dms.network.queryPeer()` - This function sends the request to lock resources to the the compute provider DMS. It takes `peerID` of the compute provider and `dms.orchestrator.bidRequest` as inputs.

`dms.orchestrator.accept()` - This function decides whether to accept the job request. It takes `dms.orchestrator.bidRequest` as input and returns a `bool` value.

`dms.orchestrator.lockResources()` - This function locks the necessary resources required for the job. It takes `dms.jobs.jobDescription` as input and returns `dms.orchestrator.bid`.

`dms.database.saveEvent()` - This function saves the event to the local database. Input value is TBD.

**List of relevant data types**:<br/>
`dms.orchestrator.computeProvidersIndex` - This contains the data of compute providers whose bids have been received. 

`dms.orchestrator.bidRequest` - This is sent to the chosen compute provider as part of job request.

`dms.orchestrator.bid` - This is the bid that is returned by the compute provider after locking of resources for the job.

`dms.orchestrator.declineJobRequest` - Message sent to the service provider if compute provider declines the job request.

`dms.orchestrator.timeoutJobRequest` - Message sent to the service provider if timeout happens.

`dms.database.declineJobEvent` - TBD


### 4. Contract Closure

* proposed 2024-03-27; by: @kabir.kbr; @janaina.senna; @0xPravar; *

The service provider and the shortlisted compute provider verify that the counterparty is a verified entity and approved by Nunet Solutions to participate in the network. This in an important step to establish trust before any work is performed.

If job does not require any payment (Volunteer Compute), contract is generated by both Service Provider and Compute Provider DMS. This is then verified by `Contract-Database`. 

Otherwise, proof of contract needs to be received from the `Contract-Database` before start of work

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/orchestrator-package-design/stages/functional_tests/features/device-management-service/orchestrator/Contract_Closure.feature))   |
| Request payload - payment included      | [ID](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/dms/config/data/id.payload.go) |
| Request payload - payment not included      | [Contract](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/tokenomics/data/contract.payload.go) |
| Return payload - Contract Exists      | [Contract](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/tokenomics/data/contract.payload.go) |
| Return payload - No Contract      | [ErrorMessage](TBD) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/sequences/contractClosure.sequence.mermaid),[.svg]()) |

**List of relevant functions**:<br/>
`dms.orchestrator.checkPaymentInfo()` - This function takes `dms.orchestrator.bidRequest` as input and checks if payment is included. It returns a bool value.

`dms.orchestrator.generateJobContract()` - This function takes `dms.orchestrator.bidRequest` as input and creates a contract on Service Provider DMS. It returns `dms.tokenomics.contract`.

`dms.orchestrator.generateJobAgreement()` - This function takes `dms.orchestrator.bidRequest` as input and creates a agreement on Compute Provider DMS. It returns `dms.tokenomics.contract`.

`dms.network.validateProvider()` - This function takes the ID of DMS (`dms.dms.config.ID`) as input and sends it to the `Contract-Database` to check whether this DMS has a valid contract. 

`contract-database.checkKYC()` - This function checks whether the DMS had done KYC and is a verified entity with a valid ID. It takes `dms.dms.config.ID` as input and returns `bool`.

`contract-database.checkContract()` - This functions takes `dms.dms.config.ID` as input and checks if the DMS whose ID has been provided has a valid contract. It returns `dms.tokenomics.contract` as proof of contract or an error message if contract does not exist.

`dms.database.saveEvent()` - This function saves the event to the local database when contract does not exist. Input value is TBD.

**List of relevant data types**:<br/>

`dms.orchestrator.bidRequest` - This is the bid request sent propagated in the network or sent to the chosen compute provider as part of job request.

`dms.tokenomics.contract` - This contains the contract data as well as proof of contract which will mention the entity that provides trust. 

`dms.dms.config.ID` - This contains identifiers like UUID, Peer ID and DID for the DMS.

### 5. Invocation and Allocation
When the contract closure workflow is completed, both the service provider and compute provider DMS have an agreement and proof of contract with them. Then the service provider DMS will send an invocation to the compute provider DMS which results in job allocation being created. Allocation can be understood as an execution space / environment on actual hardware that enables a Job to be executed.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/orchestrator-package-design/stages/functional_tests/features/device-management-service/orchestrator/Invocation_And_Allocation.feature))   |
| Request payload     | [Invocation](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/data/invocation.payload.go) |
| Data at rest       | [Allocation](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/executor/data/allocation.payload.go) |
| Return payload      | [AllocationStartSuccess](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/data/allocationStartSuccess.payload.go) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/proposed/device-management-service/orchestrator/sequences/invocationAndAllocation.sequence.mermaid),[.svg]()) |

**List of relevant functions**:<br/>

`dms.network.sendInvocation()` - This function sends the invocation to the compute provider DMS. It takes `dms.orchestrator.invocation` as input.

`dms.executor.createAllocation()` - This function creates an allocation on the compute provider DMS. It takes `dms.orchestrator.invocation` as input and returns `dms.executor.Allocation`.

**List of relevant data types**:<br/>

`dms.orchestrator.invocation` - Invocation which is sent to the compute provider DMS. This contains job description and contract data.

`dms.executor.Allocation` - This contains identifier of the allocation being created along with its status and errors (if any).

`dms.orchestrator.allocationStartSuccess` - This is the response from the compute provider DMS to Service Provider once allocation has been created.

