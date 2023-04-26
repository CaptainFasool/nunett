![latest release version](https://gitlab.com/nunet/device-management-service/-/badges/release.svg) ![unit tests](https://gitlab.com/nunet/device-management-service/badges/develop/pipeline.svg) ![coverage](https://gitlab.com/nunet/device-management-service/badges/develop/coverage.svg)

# device-management-service


The backend of the device management app.

Note: This README is intended for developers. For end users, please check our [Wiki section](https://gitlab.com/nunet/device-management-service/-/wikis/home).

For a quick install, you can use the following script:
```
sh <(curl -sL https://inst.dms.nunet.io)
```

## Getting Started with Development

The cleanest way to setup development environment is to build a deb package out of this repository and let the installer do the work for you.

```
sudo apt install build-essential curl jq iproute2
```

### Prerequisites

To build the deb, you'd be required to install these two packages:

```
sudo snap install go
sudo apt install build-essential
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

**Note**: A [Postman collection](https://gitlab.com/nunet/device-management-service/-/snippets/2507804) is there to help you get starting with REST endpoints exploration.

