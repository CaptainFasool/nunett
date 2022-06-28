#!/bin/bash

# This script is intended for developers who are developing and testing on their local system.

rm /usr/bin/nunet-dms
rm /etc/nunet/*V2.json
systemctl stop nunet-dms
rm /etc/systemd/system/nunet-dms.service
rm /etc/systemd/system/multi-user.target.wants/nunet-dms.service
systemctl daemon-reload

killall -u nunet
groupdel nunet
userdel nunet
