#!/bin/bash

# This script is intended for developers who are developing and testing on their local system.

systemctl stop nunet-dms
rm -rf /usr/bin/nunet
rm -rf /etc/nunet/*V2.json
rm -rf /etc/nunet/nunet.db
rm -rf /etc/nunet/sockets/
rm -rf /etc/systemd/system/nunet-dms.service
rm -rf /etc/systemd/system/multi-user.target.wants/nunet-dms.service
systemctl daemon-reload

killall -u nunet
groupdel nunet
userdel nunet
