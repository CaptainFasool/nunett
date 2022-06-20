# device-management-service

Backend of device management app.

## Getting Started

For getting started with development, simply install the packages:

    go install

and then run the main.go

    sudo go run main.go

Notice that I'm using `sudo` as the onboarding process writes file to `/etc/nunet`.

## Operations/Endpoints

All the endpoints metioned below should be prefixed with `/api/v1`.
### Provisioned Capacity

`GET /provisioned` endpoint returns some info about the host machine, which will be used to decide how much of capacity can be given to NuNet. See `POST /onboard`.

```
$ curl -s localhost:9999/api/v1/provisioned | jq .
{
  "cpu": 32800,
  "memory": 15843,
  "total_cores": 8
}
```

### Create a new Wallet Address

`GET /address/new` returns a wallet address and a private key. Keep the private key safe.

An example response looks like this:

```json
[
    "0x436454984F2efdDcB15f98edCEfcFc2336a2d0AF",
    "0xa9240a21bb09fed7591967ede57d53f449791b66e8ca3f8198e3f2fd037df596"
]
```

First element is the wallet address and second is the private key.

### Onboard Current Machine

`POST /onboard` onboards a new machine. Right now it onboards a machine which the service is running on.

#### Prerequisites

There are a few things you need to know before you try to onboard.

0. First, hit `GET /provisioned` to know how much `cpu` and `memory` you machine is equipped with. You'll be using this info in the JSON body of `POST /onboard` endpoint.

1. **Channel**: We currently have two channels, `nunet-development` and `nunet-private-alpha`. The former one has the cutting edge verion of the system. `nunet-private-alpha` is one of the milestore channel.

2. **Payment Address**: This is you ethereum wallet address. If you don't have an existing adderss, you can use that, or you can use the `GET /address/new` endpoint to generate a new Ethereum mainnet address.

3. **Cardano**: This one is optional, and indicates whether you want your device to be a Cardano node. Please note that, to be eligible for this, you must onboard at least of *10000MB* of memory and *6000MHz* of compute capacity.

A typical body would look like this:

```json
{
    "memory": 4000,
    "cpu": 4000,
    "channel": "nunet-private-alpha",
    "payment_addr": "0x0541422b9e05e9f0c0c9b393313279aada6eabb2",
    "cardano": false
}
```
