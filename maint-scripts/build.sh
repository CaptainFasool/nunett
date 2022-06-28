#!/bin/bash

# Build process comprises of following steps:

# Supported architectures: arm64, amd64

# INSTALLATION PROCESS

# 1. Install Nomad and Docker.
# 2. Create a nunet user. This user will be used to run the device-management-service. This user will have access to write to /etc/nunet.
# 3. Get nomad-client.service and nunet-dms.service. Copy it to appropriate location. Make sure they are run at the end of the installation process.

# UNINSTALLATION PROCESS
# 1. Stop services and remove service files.


# Both INSTALLATION PROCESS and UNINSTALLATION PROCESS are done by the package created in the build process.

# gcc is required to build go-ethereum

projectRoot=$(pwd)
outputDir="$projectRoot/dist"
version=0.1.0  # this should be dynamically set

for arch in amd64 arm64
do
    # echo .deb file will be written to: $outputDir
    archDir=$projectRoot/maint-scripts/nunet-dms_$version\_$arch
    cp -r $projectRoot/maint-scripts/nunet-dms $archDir
    sed -i "s/Version:.*/Version: $version/g" $archDir/DEBIAN/control
    sed -i "s/Architecture:.*/Architecture: $arch/g" $archDir/DEBIAN/control
    env GOOS=linux GOARCH=$arch go build -o $archDir/usr/bin/nunet-dms

    dpkg-deb --build --root-owner-group $archDir $outputDir
done



