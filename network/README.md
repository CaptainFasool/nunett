# Introduction

This package contains all network related code such as p2p communication, ip over libp2p and other networks that might be needed in the future

## Interfaces and types

### Interfaces

**Network**:

```go
// Network defines an interface provided by DMS to be implemented by various
// network providers. This could be libp2p, or other p2p providers.
type Network interface {
	// Config sets the configuration for the network in
	Config() error

	// Init initializes the node with config specific in Config() phase.
	// tags: Start
	Init() error

	// EventRegister sets handlers to handle events such as change of local address
	EventRegister() error

	// Dial connects to a peer
	// tags: ConnectPeer
	Dial() error

	// Listen listens on a connection. Example could be a stream for libp2p connection.
	Listen() error

	// Status returns status of current host in regards to implementation.
	// Output must follow a generic struct.
	Status() bool

	// Tears down network interface.
	Stop() error
}
```

Let's have a look at the methods and rationale behind them:

1. `Config()`:

`Config` is where host is prepared with desired settings. Settings are loaded from the file if required. An example in libp2p implementation would be to configure parameters which needs to be passed to `libp2p.New()` method, it can also a good place to set the stream handlers.

Things like private network are configured at this point.

2. `Init()`:

`Init` or `Start` starts up the host. This is a good place to start discovery and starting goroutines for fetching DHT update and updating peerstore.

3. `EventRegister()`:

In libp2p, we can listen to specific events. `EventRegister` is to set handler to such event. A few events in libp2p are:

- EvtLocalAddressesUpdated
- EvtLocalReachabilityChanged
- EvtPeerIdentificationCompleted
- EvtNATDeviceTypeChanged
- EvtPeerConnectednessChanged

More can be found at: <https://github.com/libp2p/go-libp2p/tree/master/core/event>

4. `Dial()`:

`Dial` is for connecting to a peer. When peer A dials peer B, peer B has to `Listen` to the incoming connection.

5. `Listen()`: 

`Listen` is counterpart to `Dial`. Read more about listening and dialing here: <https://docs.libp2p.io/concepts/transports/listen-and-dial/>

6. `Status()`:

TBD

- All peers we are corrently connected to.
- ??

7. `Stop()`: 

Basically it's like shutting down the peer. It is opposite of `Init` or `Start`.

**VPN**

```go
type VPN interface {
	// Start takes in an initial routing table and starts the VPN.
	Start() error

	// AddPeer is for adding the peers after the VPN has been started.
	// This should also update the routing table with the new peer.
	AddPeer() error

	// RemovePeer is oppisite of AddPeer. It should also update the routing table.
	RemovePeer() error

	// Stop tears down the VPN.
	Stop() error
}
```

Let's have a look at the methods and background behind them:

TBD: Parameter are still to be defined. Should we pass the peer ID? Or make it more generic to have IP addresses?

1. `Start()`:

`Start()` takes in initial list of hosts and assigns each peer a private IP.

2. `AddPeer()`:

`AddPeer()` is for adding a new peer to the VPN after the VPN is created. This should also update the routing table with the new peer. It should also not affect the existing peers, and should not lead to any IP collision.

3. `RemovePeer()`:

`RemovePeer()` should remove a peer from remove peers from the private network.

4. `Stop()`:

TBD: What should be done when the VPN is stopped? Should all the peers be removed from the network?

## Proposed Internal APIs

1. 
```
dms.network.queryPeer(dms.dms.config.peerId, dms.jobs.jobDescription) -> dms.orchestrator.computeProvidersIndex.data
```

2. 
```
dms.network.publishJob(dms.orchestrator.bidRequest(dms.jobs.jobDescription))
```

3. There should be a top level package for set of functions/initlializers for management of:

- TUN/TAP device
- Virtual ethernet
- Virtual switch
- Firewall management

## Flow of job between DMSes

Note: To be updated.

<!-- Create a sequence diagram in respective repo and link here -->

Rest API > Orchestrator > Network > Libp2p > Network > Orchestrator > Executor


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