# device-management-service

The backend of the device management app.

## Getting Started with Development

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

## Onboarding Operations/Endpoints

Refer to [wiki/Onboarding](https://gitlab.com/nunet/device-management-service/-/wikis/Onboarding)

## VM Management

Refer to [wiki/VM-Management](https://gitlab.com/nunet/device-management-service/-/wikis/VM-Management)
