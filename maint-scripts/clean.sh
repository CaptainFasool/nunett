#!/bin/bash

# This script is intended for developers who are developing and testing on their local system.

rm -rf /usr/bin/nunet-dms
rm -rf /etc/nunet/*V2.json
systemctl stop nunet-dms
systemctl stop nomad-client
rm -rf /etc/systemd/system/nunet-dms.service
rm -rf /etc/systemd/system/multi-user.target.wants/nunet-dms.service
rm -rf /etc/systemd/system/nomad-client.service
rm -rf /etc/systemd/system/multi-user.target.wants/nomad-client.service
systemctl daemon-reload

killall -u nunet
groupdel nunet
userdel nunet
