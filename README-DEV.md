Device Management Service, or DMS is the backend of all the clients which are available for compute or service provider. Currently they are SPD or Service Provider Dashboard and CPD aka Compute Provider Dashboard.

Note: This README is intended for developers. For end users, please check our [Wiki section](https://gitlab.com/nunet/device-management-service/-/wikis/home).

For a quick install, you can use the following script:
```
sh <(curl -sL https://inst.dms.nunet.io)
```

## Getting Started with Development

The cleanest way to setup development environment is to build a deb package out of this repository and let the installer do the work for you.

```
sudo apt install build-essential curl jq iproute2 libsystemd-dev
```

### Prerequisites

To build the deb, you'd be required to install these two packages:

```
sudo snap install go
sudo apt install build-essential libsystemd-dev
```

### Build, Install & Setup Dev Env

We provide a dev-setup shell script to ease the process of getting started. It does the following:

1. Setup `pre-commit` hook which runs test before every commit.
2. Build .deb file and installs it.
3. Stops the `nunet-dms` service so that we can run main.go directly.

Run the script as follows:

```
bash maint-scripts/dev-setup.sh
```

Once the env is setup, run the DMS as follows:

```
sudo go run main.go
```

Notice we're using `sudo` as the onboarding process writes some configuration files to `/etc/nunet`.

### Onboarding

You don't necessarily need to onboard for development, but that depends which part you're working on.
To onboard during development, /etc/nunet need to be manually created since it's created with the package
during installation.

Onboarding instructions can be found at [Onboarding Wiki](https://gitlab.com/nunet/device-management-service/-/wikis/Onboarding)

## Usecase: ML on GPU/CPU
### Communication between SPD and DMS

This section describes the communication mechanism between the Service Provider Daemon (SPD) and the Device Management Service (DMS). It covers the REST endpoints and Websocket interactions between the two components. 

The REST endpoints can be explored using the provided Postman collection, and any questions or clarifications can be addressed on the project's isssues. 

Websocket communication is utilized to maintain an open connection between SPD and DMS, allowing real-time exchange of messages. Message types, referred to as "actions," are used to ensure a predictable system for building and programming. The documentation explains the various actions supported by the Websocket endpoint, such as send-status, job-submitted, deployment-response, job-failed, and job-completed, along with their corresponding message formats.
#### REST Endpoints

A [Postman collection](https://gitlab.com/nunet/device-management-service/-/snippets/2507804) is there to help you get starting with REST endpoints exploration. Head over to project's issue section and create an issue with your question.

#### Websocket

Both SPD and DMS endpoints are always open through Websocket. And as we know websocket is basically a wrapper around TCP, when sending a message, it's a good practice to include the message type. This message type helps us make predictable system upon which we can build/program our system on.

We tried to write most of the logic in REST as more people are more aware to REST than websocket. But when it comes to websocket, here are the message types which we have implemented. By convention, we call message type **action**.

There is currently one websocket endpoint, which is available on `/run/deploy`, and it is available for any SPD client.

* `send-status`: This action is supposed to be invoked by SPD after they have initially invoked the `/run/request-service`. Is is called *`send-status`* because SPD is supposed to inform DMS about the blockchain transaction status of fund transfer that service provider made for running the job.

An example of send status is like this:

```json
{
    "action": "send-status",
    "message": {
        "transaction_type": "fund",
        "transaction_status": "success"
    }
}
```

`success` or `pending` indicates actual status of blockchain transaction.

send-status is the only action which a websocket client sends. Remaining others are sent by DMS.

* `job-submitted`: If you have gone through the REST endpoint, you know that first endpoint that SPD hits, is /run/request-service. Then we do the send-status. If status is success, the deployment request is then passed to the compute provider. job-submitted is sent to SPD just before sending the deployment request to compute provider.

Example:

```json
{
    "action": "job-submitted"
}
```

* `deployment-response`: Look at the message schema below.

```json
{
    "action": "deployment-response",
    "message": {
        "success": true,
        "content": https://log.url/user/logid
    }
}
```

If deployment is successful on the compute provider side, SPD will receive this message with log URL.

* `job-failed`: If somehow job was not able to start, due to any error, SPD is going to receive following message:

```json
{
    "action": "job-failed",
    "message": "reason why deployment failed"
}
```

* `job-completed`: This action is sent when container has exited with 0 exit status. Message looks simple similar to job-submitted.

```json
{
    "action": "job-completed"
}
```

## Run 2 DMS side by side

As a developer, you might be in a situation where you have to both service provider and compute provider. For that you'll have to run 2 DMS, one acting as SP and another CP.

**Step 1**:

Clone the repo to 2 different directory. You might want to use descriptive directory names to avoid confusion.

**Step 2**:

You need to modify some configuration so that both the DMS does not create a deadlock. Those configuration include:

1. The port DMS is listening on.
2. The database file DMS uses.

dms_config.json can be used to modify those settings. Here is a sample config file which can be modified to your taste and used:

```json
{
  "p2p": {
    "listen_address": [
      "/ip4/0.0.0.0/tcp/9100",
      "/ip4/0.0.0.0/udp/9100/quic"
    ]
  },
  "general": {
    "metadata_path": "/home/santosh/.config/nunet/dms/",
    "debug": true
  },
  "rest": {
    "port": 10000
  }
}
```

Please use absolute paths to keep yourself out of trouble. Moreover, have a look at [config structure](https://gitlab.com/nunet/device-management-service/-/blob/develop/internal/config/config.go).

3. You must also change the port number in the nunet shell script if you are planning to use nunet cli.

**Step 3**:

Onboard both DMSes.

**Step 4**:

Check if both can discover each other.

**Step 5**:

Change DMS backend URL from the SPD/CPD side and start with the testing.

## Run Security Test Suite

This command will run the Security Test suite:
`go test -ldflags="-extldflags=-Wl,-z,lazy" -run=TestSecurity`

## License

Device Management Service (DMS) is licensed under the [GNU AFFERO GENERAL PUBLIC LICENSE](https://www.gnu.org/licenses/agpl-3.0.txt).

