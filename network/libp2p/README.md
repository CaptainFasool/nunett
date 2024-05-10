# Introduction

This package implements `Network` interface defined in root level network dir.

## Requirements

proposed: @sam

### Peer discovery and handshake

Instead of keeping track of all the peers. Peers should only in touch with peers of their types in terms of network latency, resources, or uptime.

A reason for this is, if some low performing peer is with some high performing peers, and job is distributed among them, it can slow others peers as well overall. 

#### Max number of handshake peers

Different nodes will have different requirements regarding the number of peers that they should remain handshaking with. e.g. a small node on a mobile network will not need to maintain a large list of peers. But, a node acting as network load balancer in a data center might need to maintain a large list of peers.

#### Filter list

We can have filters that ensures that the only peers that are handshaked with are ones that meet certain criteria. The following list is not exhaustive:

1. Network latency. Have a set of fastest peers.
2. Resource. Relates to job types.
3. Uptime. Connect to peers who are online for certain period of time.

#### Network latency

For the network latency part, DMS should also be able to keep latency table between ongoing jobs on different CPs. The network package should be able to report it to the master node (SP). Orchestrator can then make decision on whether to replace workers or not.

**Thoughts**:

* Filter peers at the time of discovery. Based on above parameters.
* SP/orchestrator specifies what pool of CP it is looking for.
* CP connects to same kind of CP.
* Can use gossipsub.

## libp2p package as it is now

Note: The section will be deprecated after refactor. We may update the section, or remove it completely if it does not adds any value.

### Current (sub)packages:

- `libp2p` - the core of DMS to DMS communication, more on this later
- `libp2p/machines` - mostly job handling on the SP side e.g. filtering, job request handler
- `libp2p/pubsub` - current pubsub implementation
- `firecracker/networking` - creates and configures tap device
- `internal` - deals with websocket client
- `internal/messaging`? - workers working on job queues

### Functionalities

As soon as DMS starts, and if it is onboarded to the network, `libp2p.RunNode` is executed. This gets up entire thing related to libp2p. Let's run down through it to see what it does.

1. RunNode calls `NewHost`. NewHost in itself does a lot of things. Let's dive into the NewHost:

- It creates a [connection manager](https://github.com/libp2p/go-libp2p/blob/v0.33.2/p2p/net/connmgr/connmgr.go#L110). This defines what is the upper and lower limit of peers current peer will connect to.
- It then defines a multiaddr filter which is used to deny discovering on local network. This was added to stop scanning local network in a data center.
- NewHost then sets various options for the and passes it to libp2p.New to create a new host. Options such as NAT traversal is configured at this point.

2. Getting back to other RunNode, it calls `p2p.BootstrapNode(ctx)`. Bootstrapping basically is connecting to initial peers.

3. Then the function continues setting up streams. Streams are bidirectional connection between peers. More on this in next section. Here is an example of setting up a stream handler on host for particular protocol:

```go
host.SetStreamHandler(protocol.ID(DepReqProtocolID), depReqStreamHandler)
```

4. After that, we have `discoverPeers`, 


```go
go p2p.StartDiscovery(ctx, utils.GetChannelName())
```

5. After that, we have DHT update and get functions to store information about peer in peerstore.


## Streams

Communication between libp2p peers, or more generally DMS happens using libp2p streams. A DMS can have one or many stream with one or more peer. We currently we have adopted following streams for our usecases.

1. Ping

We can count this as internal to libp2p and is used for operational purposes. Unlike ICMP pings, libp2p [pings works on streams](https://github.com/libp2p/specs/blob/master/ping/ping.md), and is closed after the ping.

2. Chat

A utility functionality to enable chat between peers.

3. VPN

Most recent addition to DMS, where we send IP packets through libp2p stream.

4. File Transfer

File transfer is generally used to carry files from one DMS to another. Most notably used to carry checkpoint files from a job from CP to SP.

5. Deployment Request (DepReq)

Used for deployment of a job and for getting their progress.


### Current DepReq Stream Handler

Each stream need to have a handler attached to it. Let's get to know more about **deployment request** stream handler. Deployment request handler handles incoming deployment request from the service provider side. Similarly, some function has to listen for update from the service provider side as well. More on that in the next in a minute.

Following is a sequence of event happening on compute provider side:

1. Checks if `InboundDepReqStream` variable is set. And if it is, reply to service provider: `DepReq open stream length exceeded`. Currently we have only 1 job allowed per dep req stream.
2. If above is not the case, we go and read from the stream. We are expecting a 
3. Now is the point to set the `InboundDepReqStream` to the incoming stream value.
4. In unmarshal the incoming message into `models.DeploymentRequest` struct. If it can't, it informs the other party about the it.
5. Otherwise, if everything is going well till now, we check the `txHash` value from the depReq. And make sure it exist on the blockchain before proceeding. If the txHash is not valid, or it timed out while waiting for validation, we let the other side know.
6. Final thing we do it to put the depReq inside the `DepReqQueue`.

After this step, the command is handed over to executor module. Please refer to [Relation between libp2p and docker modules](#relation-between-libp2p-and-docker-modules-now-executor).

Deployment Request stream handler can further be segmented into different message types:

```
MsgDepReq    = "DepReq"
MsgDepResp   = "DepResp"
MsgJobStatus = "JobStatus"
MsgLogStderr = "LogStderr"
MsgLogStdout = "LogStdout"
```
Above message types are used by various functions inside the stream. Last 4 or above is handled on the SP side. Further by the websocket server which started the deployment request. This does not means CP does not deals with them. Keep reading.

Actual data model is described in current data models section below.

1. **DeploymentResponse**: DeploymentResponse is initial response from the CP to SP. It tells the SP that if deployment was successful or was declined due to operational or validational reasons. Most of the validation is just error check at stream handling or executor level.
2. **DeploymentUpdate**: DeploymentUpdate update is used to inform SP about the state of the job. Most of the update is handled using libp2p stream on network level and websocket on the user level. There is no REST API defined. This should change in next iteration. See the proposed section for this.

On the service provider side, we have `DeploymentUpdateListener` listening to the stream for any activity from the computer provider for update on the job. 

Based on the message types, it does specific actions, which is more or less sending it to websocket client. These message types are `MsgJobStatus`, `MsgDepResp`, `MsgLogStdout` and `MsgLogStderr`

## Relation between libp2p and docker modules

When DepReq streams receives a deployment request on the stream, it does some json validation, and pushes it to `DepReqQueue`. This extra step instead of directly passing the command to docker package was for decoupling and scalibility.

There is a `messaging.DeploymentWorker()` goroutine which is launched at DMS startup in `dms.Run()`.

This `messaging.DeploymentWorker()` is the crux of the job deployment, as what is done in current proposed version of DMS. Based on executor type (currently firecracker and docker), it was passed to specific functions on different modules.
