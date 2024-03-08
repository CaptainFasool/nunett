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

### Device Endpoints

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
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/DeviceStatusHandler.sequence.mermaid?ref_type=heads),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/DeviceStatusHandler.sequence.svg?ref_type=heads)) | 

#### Change Device Status

**endpoint**: `/device/status`<br/>
**method**: `HTTP POST`<br/>
**output**: `Success Message`

The endpoint changes the current status of the machine.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/device/Change_Device_Status.feature))   |
| Request payload       | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/ChangeDeviceStatusHandler.sequence.mermaid?ref_type=heads),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/ChangeDeviceStatusHandler.sequence.svg?ref_type=heads)) | 

### Onboarding endpoints
These endpoints are related to the onboarding functionality of DMS.

#### Create Payment Address

**endpoint**: `/onboarding/address/new`<br/>
**method**: `HTTP GET`<br/>
**output**: `Public-Private key pair & Mnemonic`

This endpoint creates a new blockchain payment address for the user.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/onboarding/Create_Payment_Address.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/CreatePaymentAddressHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/CreatePaymentAddressHandler.sequence.svg)) | 

#### Onboard

**endpoint**: `/onboarding/onboard`<br/>
**method**: `HTTP POST`<br/>
**output**: `Machine Metadata`

This endpoint executes the onboarding process for a compute provider device.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/onboarding/Onboard.feature))   |
| Request payload       | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/OnboardHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/OnboardHandler.sequence.svg)) | 

#### Get Metadata

**endpoint**: `/onboarding/metadata`<br/>
**method**: `HTTP GET`<br/>
**output**: `Machine Metadata`

This endpoint fetches the current metadata of the onboarded device.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/onboarding/Get_Metadata.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/GetMetadataHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/GetMetadataHandler.sequence.svg)) | 

#### Provisioned Capacity

**endpoint**: `/onboarding/provisioned`<br/>
**method**: `HTTP GET`<br/>
**output**: `Provisioned Capacity`

This endpoint fetches the total capacity of the machine that is onboarded to Nunet.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/onboarding/Provisioned_Capacity.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/ProvisionedCapacityHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/ProvisionedCapacityHandler.sequence.svg)) | 

#### Onboard Status

**endpoint**: `/onboarding/status`<br/>
**method**: `HTTP GET`<br/>
**output**: `Onboarding status & Metadata`

This endpoint returns onboarding status of the machine along with some metadata.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/onboarding/Onboard_Status.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/OnboardStatusHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/OnboardStatusHandler.sequence.svg)) | 

#### Resource Config

**endpoint**: `/onboarding/resource-config`<br/>
**method**: `HTTP POST`<br/>
**output**: `Machine Metadata`

This endpoint allows the user to change the configuration of the resources onboarded to Nunet.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/onboarding/Resource_Config.feature))   |
| Request payload       | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/ResourceConfigHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/ResourceConfigHandler.sequence.svg)) | 

#### Offboard

**endpoint**: `/onboarding/offboard`<br/>
**method**: `HTTP DELETE`<br/>
**output**: `Success Message`

This endpoint allows the user to remove the resources onboarded to Nunet. 

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/onboarding/Offboard.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/OffboardHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/OffboardHandler.sequence.svg)) | 

### Peers Endpoints

#### List Peers

**endpoint**: `/peers`<br/>
**method**: `HTTP GET`<br/>
**output**: `Peer List`

This endpoint gets a list of peers that the node can see within the network.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/List_Peers.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/ListPeersHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/ListPeersHandler.sequence.svg)) | 

#### List DHT Peers

**endpoint**: `/peers/dht`<br/>
**method**: `HTTP GET`<br/>
**output**: `DHT Peer List`

This endpoint gets a list of peers that the node has received a dht update from.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/List_DHT_Peers.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/ListDHTPeersHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/ListDHTPeersHandler.sequence.svg)) | 

#### List Kad DHT Peers

**endpoint**: `/peers/kad-dht`<br/>
**method**: `HTTP GET`<br/>
**output**: `DHT Peer List`

This endpoint gets a list of peers that the node has received a dht update from.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/List_KadDHT_Peers.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/ListKadDHTPeersHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/ListKadDHTPeersHandler.sequence.svg)) | 

#### Self Peer Info

**endpoint**: `/peers/self`<br/>
**method**: `HTTP GET`<br/>
**output**: `Peer Info`

This endpoint gets the peer info of the libp2p node.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/Self_Peer_Info.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/SelfPeerInfoHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/SelfPeerInfoHandler.sequence.svg)) | 

#### List Chat

**endpoint**: `/peers/chat`<br/>
**method**: `HTTP GET`<br/>
**output**: `List of chats`

This endpoint gets the list of chat requests from peers.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/List_Chat.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Error logged to ELasticsearch      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/ListChatHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/ListChatHandler.sequence.svg)) | 

#### Clear Chat

**endpoint**: `/peers/chat/clear`<br/>
**method**: `HTTP GET`<br/>
**output**: `Message`

This endpoint clears the chat request streams from peers.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/Clear_Chat.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Error logged to ELasticsearch      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/ClearChatHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/ClearChatHandler.sequence.svg)) | 

#### Start Chat

**endpoint**: `/peers/chat/start`<br/>
**method**: `HTTP GET`<br/>
**output**: `None`

This endpoint starts a chat session with a peer.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/Start_Chat.feature))   |
| Request payload       | None |
| Return payload - success     | None |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/StartChatHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/StartChatHandler.sequence.svg)) | 

#### Join Chat

**endpoint**: `/peers/chat/join`<br/>
**method**: `HTTP GET`<br/>
**output**: `None`

This endpoint allows the user to join a chat session started by a peer.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/Join_Chat.feature))   |
| Request payload       | None |
| Return payload - success     | None |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/JoinChatHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/JoinChatHandler.sequence.svg)) | 

#### Dump DHT

**endpoint**: `/dht`<br/>
**method**: `HTTP GET`<br/>
**output**: `DHT Content`

This endpoint returns the entire DHT content.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/Dump_DHT.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/DumpDHTHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/DumpDHTHandler.sequence.svg)) | 

#### Default DepReq Peer

**endpoint**: `/peers/depreq`<br/>
**method**: `HTTP GET`<br/>
**output**: `Message including peerID`

This endpoint is used to set peer as the default receipient of deployment requests by setting the peerID parameter on GET request. 

Note:
* By sending a GET request without any parameters we get the peer currently set as default deployment request receiver. 

* Sending a GET request with `peerID` parameter set to '0' will remove default deployment request receiver.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/Default_DepReq_Peer.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/DefaultDepReqPeerHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/DefaultDepReqPeerHandler.sequence.svg)) | 

#### Clear File Transfer Requests

**endpoint**: `/peers/file/clear`<br/>
**method**: `HTTP GET`<br/>
**output**: `Message`

This endpoint is used to clear file transfer request streams from peers.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/Clear_File_Transfer_Requests.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/ClearFileTransferRequestsHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/ClearFileTransferRequestsHandler.sequence.svg)) | 

#### List File Transfer Requests

**endpoint**: `/peers/file`<br/>
**method**: `HTTP GET`<br/>
**output**: `Message`

This endpoint is used to get a list of file transfer requests from peers.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/List_File_Transfer_Requests.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/ListFileTransferRequestsHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/ListFileTransferRequestsHandler.sequence.svg)) | 

#### Send File Transfer

**endpoint**: `/peers/file/send`<br/>
**method**: `HTTP GET`<br/>
**output**: `NIL`

This endpoint is used to initiate file transfer to a peer. Note that `filePath` and `peerID` are required arguments.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/Send_File_Transfer.feature))   |
| Request payload       | None |
| Return payload - success     | `NIL` |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/SendFileTransferHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/SendFileTransferHandler.sequence.svg)) | 

#### Accept File Transfer

**endpoint**: `/peers/file/accept`<br/>
**method**: `HTTP GET`<br/>
**output**: `NIL`

This endpoint is used to initiate file transfer to a peer. Note that `filePath` and `peerID` are required arguments.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/peers/Accept_File_Transfer.feature))   |
| Request payload       | None |
| Return payload - success     | `NIL` |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/AcceptFileTransferHandler.sequence.mermaid,[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/AcceptFileTransferHandler.sequence.svg)) | 

### Run Endpoints

#### Request Service

**endpoint**: `/run/request-service`<br/>
**method**: `HTTP POST`<br/>
**output**: `Funding Response`

This endpoint searches the DHT for non-busy, available devices with appropriate metadata. Then informs parameters related to blockchain to request to run a service on NuNet.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/run/Request_Service.feature))   |
| Request payload       | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/RequestServiceHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/RequestServiceHandler.sequence.svg)) | 

#### Deployment Request

**endpoint**: `/run/deploy`<br/>
**method**: `HTTP GET`<br/>
**output**: `None`

This endpoint loads deployment request from the database after a successful blockchain transaction has been made and passes it to compute provider.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/run/Deployment_Request.feature))   |
| Request payload       | None |
| Return payload - success     | None |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/DeploymentRequestHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/DeploymentRequestHandler.sequence.svg)) | 

#### List Checkpoint

**endpoint**: `/run/checkpoints`<br/>
**method**: `HTTP GET`<br/>
**output**: `Checkpoints`

This endpoint scans lists all the files which can be used to resume a job. Returns a list of objects with absolute path and last modified date.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/run/List_Checkpoint.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/ListCheckpointHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/ListCheckpointHandler.sequence.svg)) | 

### Telemetry Endpoints

#### Get Free Resources

**endpoint**: `/telemetry/free`<br/>
**method**: `HTTP GET`<br/>
**output**: `Free resources`

This endpoint checks and returns the amount of free resources available in a machine.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/telemetry/Get_Free_Resources.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/GetFreeResourcesHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/GetFreeResourcesHandler.sequence.svg)) |

### Transactions Endpoints

#### Get Job Transaction Hashes

**endpoint**: `/transactions`<br/>
**method**: `HTTP GET`<br/>
**output**: `Transaction Hashes`

This endpoint gets the list of transaction hashes along with the date and time of jobs done.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/transactions/Get_JobTx_Hashes.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/GetJobTxHashesHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/GetJobTxHashesHandler.sequence.svg)) |

#### Request Reward

**endpoint**: `/transactions/request-reward`<br/>
**method**: `HTTP POST`<br/>
**output**: `Reward Response`

This endpoint takes request from the compute provider, talks with Oracle and releases tokens if conditions are met.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/transactions/Request_Reward.feature))   |
| Request payload       | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/RequestRewardHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/RequestRewardHandler.sequence.svg)) |

#### Send Transaction Status

**endpoint**: `/transactions/send-status`<br/>
**method**: `HTTP POST`<br/>
**output**: `Transaction Status`

This endpoint returns the status of a blockchain transaction such as token withrawal.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/transactions/Send_Tx_Status.feature))   |
| Request payload       | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/SendTxStatusHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/SendTxStatusHandler.sequence.svg)) |

#### Update Transaction Status

**endpoint**: `/transactions/update-status`<br/>
**method**: `HTTP POST`<br/>
**output**: `Message`

This endpoint updates the status of saved transactions by fetching info from blockchain using koios REST API.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/transactions/Update_Tx_Status.feature))   |
| Request payload       | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/UpdateTxStatusHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/UpdateTxStatusHandler.sequence.svg)) |

### VM Endpoints

#### Start Custom

**endpoint**: `/vm/start-custom`<br/>
**method**: `HTTP POST`<br/>
**output**: `Message`

This endpoint start a firecracker VM with custom configuration.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/vm/Start_Custom.feature))   |
| Request payload       | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/StartCustomHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/StartCustomHandler.sequence.svg)) |

#### Start Default

**endpoint**: `/vm/start-default`<br/>
**method**: `HTTP POST`<br/>
**output**: `Message`

This endpoint start a firecracker VM with default configuration.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/vm/Start_Default.feature))   |
| Request payload       | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/StartDefaultHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/StartDefaultHandler.sequence.svg)) |

### Debug Endpoints

These endpoints are only available when `DEBUG` mode is enabled.

#### Manual DHT Update

**endpoint**: `/dht/update`<br/>
**method**: `HTTP GET`<br/>
**output**: `Message`

This endpoint initiates a manual update of the DHT.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/debug/Manual_DHT_Update.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/ManualDHTUpdateHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/ManualDHTUpdateHandler.sequence.svg)) |

#### Cleanup Peer

**endpoint**: `/cleanup`<br/>
**method**: `HTTP GET`<br/>
**output**: `Message`

This endpoint removes a peer from the local database.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/debug/Cleanup_Peer.feature))   |
| Request payload       | None |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/CleanupPeerHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/CleanupPeerHandler.sequence.svg)) |

#### Ping Peer

**endpoint**: `/ping`<br/>
**method**: `HTTP GET`<br/>
**output**: `Ping Peer Response`

This endpoint pings a peer and checks the peer's presence in the DHT. It also returns the round trip time (RTT) for the ping.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/debug/Ping_Peer.feature))   |
| Request payload       | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/PingPeerHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/PingPeerHandler.sequence.svg)) |

#### Old Ping Peer

**endpoint**: `/oldping`<br/>
**method**: `HTTP GET`<br/>
**output**: `Ping Peer Response`

This endpoint pings a peer and checks the peer's presence in the DHT. It also returns the round trip time (RTT) for the ping.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/debug/Old_Ping_Peer.feature))   |
| Request payload       | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/OldPingPeerHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/OldPingPeerHandler.sequence.svg)) |

#### Dump Kademlia DHT

**endpoint**: `/kad-dht`<br/>
**method**: `HTTP GET`<br/>
**output**: `DHT Content`

This endpoint returns the DHT contents.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/dms-rest-api/stages/functional_tests/features/device-management-service/api/debug/Dump_Kademlia_DHT.feature))   |
| Request payload       | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - success     | entityDiagrams ([.mermaid](),[.svg]()) |
| Return payload - error      | entityDiagrams ([.mermaid](),[.svg]()) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/DumpKademliaDHTHandler.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/dms-rest-api/device-management-service/api/sequences/rendered/DumpKademliaDHTHandler.sequence.svg)) |



