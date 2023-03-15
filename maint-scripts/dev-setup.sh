#!/bin/sh

# setupRepo configures pre-commit hook
function setupRepo() {
    cp maint-scripts/pre-commit .git/hooks/pre-commit
}

function installPreReqUbuntu() {
    bash maint-scripts/build.sh && sudo apt install dist/*deb
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
