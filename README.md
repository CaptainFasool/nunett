# device-management-service

The backend of the device management app.

Note: This README is intended for developers. For end users, please check our [Wiki section](https://gitlab.com/nunet/device-management-service/-/wikis/home).

## Getting Started with Development

Operations done via Device Management Service (DMS) depends on packages such as:

- docker
- iptables
- ip

For end users, these are installed as part of .deb package. these prerequisites will be taken care of, otherwise, you need to install them manually.

### Setup Development Environment

On Debian-based systems, this command should get you ready with the installation.

```
sudo apt install build-essential curl golang jq
```

Similarly, on RHEL based system, the equivalent command would be:

```
sudo yum group install "Development Tools"
sudo yum install curl golang jq
```

Please make sure you have the appropriate Go version installed. We work with and test against the latest Go release. You can find the version we are using in the go.mod file.

### Build and Run the server


If you have Go installed, next install the packages:

    go install

and then run the main.go

    sudo go run main.go

Notice that I'm using `sudo` as the onboarding process writes some configuration files to `/etc/nunet`.

Note about firecracker VMs. DMS also depends on binaries such as `firecracker` and one custom build binary, source code of which is stored in ./maint-script directory. Store them somewhere in $PATH.


## Onboarding Operations/Endpoints

Refer to [wiki/Onboarding](https://gitlab.com/nunet/device-management-service/-/wikis/Onboarding)

## VM Management

Refer to [wiki/VM-Management](https://gitlab.com/nunet/device-management-service/-/wikis/VM-Management)
