## Introduction
This package takes care of job scheduling and management (manages jobs on other DMSs)

## Job Orchestration
The lifecyle of a job on Nunet platform consists of various operations from job posting to settlement of the contract. A key distinction to note is the option of two types of orchestration mechanisms: `push` and `pull`. Broadly speaking `pull` orchestration works on the premise that resource providers bid for jobs available in the network. 

Whereas `push` orchestration develops on the idea that users choose from the available providers and their resources. However, given the decentralised and open nature of the platform, it may be required to engage the providers to get their current (latest) state and preferences. This leads to an overlap with the `pull` orchestration approach.

The default setting is to use `pull` based orchestration. However, the user can choose to use `push` based orchestration to suit their needs.

The details of all the operations involved in the job orchestration are described in the following sections.

### 1. Job Posting
The first step is when a user posts a request to run a computing job. This should define various job requirements and preferences.

**endpoint**: `/orchestrator/postJob`<br/>
**method**: `HTTP POST`<br/>
**output**: `None`

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin]())   |
| Request payload       | [jobDescription](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/jobs/data/jobDescription.payload.go)|
| Return payload       | None |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/sequences/jobPosting.sequence.mermaid),[.svg]()) | 

List of relevant functions:<br/>
`dms.orchestrator.processJob` - This function will validate the job received, add metadata (if needed) and save the job to the local database.

List of relevant data types:<br/>
`dms.jobs.jobDescription` - This contains the job details and desired capability needed to execute the job.

### 2. Search and Match
Once the DMS received a job posting, it will look to find nodes that can service the request. This is done by matching the job requirements with the available resources.

### Configuration
`defaultOrchestrationType`: As mentioned before, the default setting of the network is use `pull` search and match operation. This value is stored in `defaultOrchestrationType` parameter saved in the [config](https://gitlab.com/nunet/device-management-service/-/tree/orchestrator-package-design/dms/config) folder under `dms` package. The user can override this value to `push` during job request.

`defaultSearchTimeout`: Each DMS will have default timeout value for the search operation. This value is stored in `defaultSearchTimeout` parameter saved in the [config](https://gitlab.com/nunet/device-management-service/-/tree/orchestrator-package-design/dms/config) folder under `dms` package. This can be overridden by the owner of the DMS.

### Pull Based
The first step is to request bids from the compute providers in the network. DMS compares the capability of the available resources against soft and hard constraints specified in the job requirements. The final outcome is a list of eligible compute providers with their bids.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin]())   |
| Request payload       | [BidRequest](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/data/bidRequest.payload.go)|
| Data at rest (CP DMS)      | [EligibleComputeProvidersData](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/data/capabilityComparison.payload.go) |
| Return payload       | [Bid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/data/bid.payload.go) |
| Data at rest (SP DMS)       | [EligibleComputeProvidersIndex](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/data/computeProviderIndex.payload.go) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/sequences/pullSearchAndMatch.sequence.mermaid),[.svg]()) |

The second step is to shortlist the preferred compute provider peer based on some selection criteria. Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin]())   |
| Request payload       | [EligibleComputeProvidersIndex](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/data/computeProviderIndex.payload.go) |
| Return payload       | [EligibleComputeProviderData](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/data/computeProviderIndex.payload.go) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/sequences/selectPreferredNode.sequence.mermaid),[.svg]()) |

### 3. Job Request
In case the shortlisted compute provider has not locked the resources while submitting the bid, the job request workflow is executed. This requires the compute provider DMS to lock the necessary resources required for the job and re-submit the bid. Note that at this stage compute provider can still decline the job request.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin]())   |
| Request payload       | [BidRequest](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/data/bidRequest.payload.go) |
| Return payload - request acceptance      | [Bid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/data/bid.payload.go) |
| Return payload - request denied      | [DeclineMessage](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/data/declineJobRequest.payload.go) |
| Return payload - timeout      | [TimeoutResponse](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/data/timeoutJobRequest.payload.go) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/sequences/jobRequest.sequence.mermaid),[.svg]()) |

### 4. Contract Closure
The service provider and the shortlisted compute provider verify that the counterparty is verified and approved by Nunet Solutions to participate in the network. This in an important step to establish trust before any work is performed. 

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin]())   |
| Request payload       | [UUID]() |
| Return payload - confirmation      | [Contract]() |
| Return payload - verification failure      | [DeclineMessage](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/data/declineJobRequest.payload.go) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/orchestrator-package-design/device-management-service/orchestrator/sequences/contractClosure.sequence.mermaid),[.svg]()) |







