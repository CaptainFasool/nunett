# Introduction

This package contains all network related code such as p2p communication, ip over libp2p and other networks that might be needed in the future

## Private Network
The private network functionality allows users to create and join a private network with some authorised peers. Note that these peers need to be identified beforehand to use this feature. It is also required that all peers have onboarded to the Nunet network and are using the same channel. It is because the identification of peers is done using libp2p public key generated during the onboarding process.

The creation of private network consists of the following operations.

### Configure Private Network
The user who wants to create the private network should have a list of `peer ids` it wants to authorise to join the private network. This process configures the network by creating a swarm key and a bootstrap node.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/private-network/stages/functional_tests/device-management-service/features/Configure_Private_Network.feature))   |
| Request payload       | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/authorisedMachines.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/authorisedMachines.payload.svg)) |
| Return payload       | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/configurationStatus.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/configurationStatus.payload.svg)) |
| Data at rest       | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/privateNetworkData.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/privateNetworkData.payload.svg)) | 
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/sequences/configurePrivateNetwork.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/sequences/rendered/configurePrivateNetwork.sequence.svg)) | 


### Swarm Key Request
The authorised user who wants to connect to the private network will request for a swarm key from the node that has configured the private network.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/private-network/stages/functional_tests/device-management-service/features/Request_Swarm_Key.feature))   |
| Request payload       | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/requestSwarmKey.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/requestSwarmKey.payload.svg)) |
| Return payload       | entityDiagrams ([.mermaid](),[.svg]()) |
| Other Data - p2p request  | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/swarmKeyRequest.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/swarmKeyRequest.payload.svg)) |
| Other Data - p2p response  | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/shareSwarmKey.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/shareSwarmKey.payload.svg)) |
| Data at rest       | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/privateNetworkInfo.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/privateNetworkInfo.payload.svg)) | 
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/sequences/exchangeSwarmKey.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/sequences/rendered/exchangeSwarmKey.sequence.svg)) | 

### Share Swarm Key
The node which has created the swarm key shares it with the authorised user when requested. The authentication of the user is based on its public key.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/private-network/stages/functional_tests/device-management-service/features/Share_Swarm_Key.feature))   |
| Request payload       | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/swarmKeyRequest.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/swarmKeyRequest.payload.svg)) |
| Return payload       | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/shareSwarmKey.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/shareSwarmKey.payload.svg)) |
| Other Data - user notification     | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/shareSwarmKeySucess.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/shareSwarmKeySucess.payload.svg)) | 
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/tree/private-network/device-management-service/network/sequences),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/tree/private-network/device-management-service/network/sequences/rendered)) | 

### Join Private Network
The DMS will disconnect from the public network and join the private network using the shared swarm key.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/private-network/stages/functional_tests/device-management-service/features/Join_Private_Network.feature))   |
| Request payload       | None |
| Return payload       | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/joinPrivateNetworkSucess.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/joinPrivateNetworkSucess.payload.svg)) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/sequences/joinPrivateNetwork.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/sequences/rendered/joinPrivateNetwork.sequence.svg)) | 

### Disconnect Private Network
The DMS will disconnect from the private network and join the public network the user onboarded on to.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/private-network/stages/functional_tests/device-management-service/features/Disconnect_Private_Network.feature))   |
| Request payload       | None |
| Return payload       | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/disconnectSucess.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/disconnectSucess.payload.svg)) |
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/sequences/disconnectPrivateNetwork.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/sequences/rendered/disconnectPrivateNetwork.sequence.svg)) |

### Rotate Swarm Key
The DMS will generate a new swarm key for the private network and notify the authorised users.

Please see below for relevant specification and data models.

| Spec type              | Location |
---|---|
| Features / test case specifications | Scenarios ([.gherkin](https://gitlab.com/nunet/test-suite/-/blob/private-network/stages/functional_tests/device-management-service/features/Rotate_Swarm_Key.feature))   |
| Request payload       | None |
| Return payload       | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rotateSwarmKeySuccess.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/rotateSwarmKeySuccess.payload.svg)) |
| Data at rest       | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/privateNetworkData.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/privateNetworkData.payload.svg)) | 
| Other Data - peer notification   | entityDiagrams ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/swarmKeyChanged.payload.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/data/rendered/swarmKeyChanged.payload.svg)) | 
| Processes / Functions | sequenceDiagram ([.mermaid](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/sequences/rotateSwarmKey.sequence.mermaid),[.svg](https://gitlab.com/nunet/open-api/platform-data-model/-/blob/private-network/device-management-service/network/sequences/rendered/rotateSwarmKey.sequence.svg)) |