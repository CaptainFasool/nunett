![latest release version](https://gitlab.com/nunet/device-management-service/-/badges/release.svg) ![unit tests](https://gitlab.com/nunet/device-management-service/badges/develop/pipeline.svg) ![coverage](https://gitlab.com/nunet/device-management-service/badges/develop/coverage.svg)

# device-management-service


The backend of the device management app.

Note: This README is intended for developers. For end users, please check our [Wiki section](https://gitlab.com/nunet/device-management-service/-/wikis/home).

For a quick install, you can use the following script:
```
curl -L https://inst.dms.nunet.io | sh
```

## Getting Started with Development

Operations done via Device Management Service (DMS) depends on packages such as:

- docker
- iptables
- ip

For end users, these are installed as part of .deb package. these prerequisites will be taken care of, otherwise, you need to install them manually.

### Setup Development Environment

You can install Go using the following commands below on both Debian or RHEL based systems.

#### Install Go based on official documentation (Debian/RHEL based system)
```
wget https://go.dev/dl/go1.19.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
source $HOME/.profile
```
#### Install additional dependencies on Debian based system

```
sudo apt install build-essential curl jq
```

#### Install additional dependencies on RHEL based system

```
sudo yum install curl jq
```

Please make sure you have the appropriate Go version installed, with the `go version` command. We work with and test against the latest Go release.

### Build and Run the server


If you have Go installed, download the develop branch from this repository:

    git clone -b develop https://gitlab.com/nunet/device-management-service.git dms && cd dms

Next, install the packages:

    go install

and then run main.go

    sudo go run main.go

Notice we're using `sudo` as the onboarding process writes some configuration files to `/etc/nunet`.

Note about firecracker VMs. DMS also depends on binaries such as `firecracker` and one custom build binary, source code of which is stored in ./maint-script directory. Store them somewhere in $PATH.


## Onboarding Operations/Endpoints

Refer to [wiki/Onboarding](https://gitlab.com/nunet/device-management-service/-/wikis/Onboarding)

## VM Management

Refer to [wiki/VM-Management](https://gitlab.com/nunet/device-management-service/-/wikis/VM-Management)
