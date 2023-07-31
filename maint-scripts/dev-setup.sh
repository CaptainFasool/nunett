#!/bin/bash

# setupRepo configures pre-commit hook
function setupRepo() {
    cp maint-scripts/pre-commit .git/hooks/pre-commit
}

function installPreReqUbuntu() {
    bash maint-scripts/build.sh
    newest=$(ls -1v dist/*.deb | tail -n1)
    sudo apt install "./$newest"
    sudo systemctl stop nunet-dms
}

if [ -f /etc/os-release ]; then
    . /etc/os-release

    if [ "$ID" == "ubuntu" ]; then
        setupRepo
        installPreReqUbuntu
    else
        echo "dev-setup.sh is work in progress and is currently only tested on Ubuntu."
        echo "send an merge request with support with your distro..."
    fi
else
    echo Not able to identify your distro/OS.
fi
