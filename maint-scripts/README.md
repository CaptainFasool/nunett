There are 3 files/subdirectory in this directory. Here are what they are for:

1. `nunet-dms/`: This is a template directory to build deb file. This direcotry is used by `build.sh` to write the binary file in the `/usr/bin`, update architecture and version in the control file. And then build a .deb file out of the direcotry.

2. `build.sh`: This script is intended to be used by CI/CD server. This script creates .deb package for `amd64` and `arm64`.

3. `clean.sh`: This script is intended to be used by developers. People who has not installed Device Management Service via package manager.

---

## How to build

### Requirements/dependencies:

- `golang` is required to build the nunet-dms binary
- `gcc` is required to build go-ethereum
- `dpkg-deb` is required to build the debian package

### Invoking a build

Build is supposed to be invoked from the root of the project.

A build can be invoked by:

    bash maint-script/build.sh
