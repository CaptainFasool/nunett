#!/bin/bash
kernel_path="/home/santosh/firecracker/vmlinux.bin"
rootfs_path="/home/santosh/firecracker/bionic.rootfs.ext4"

# Setting kernel, rootfs and resources can be done with config files. Thus excluding them. 
function set_kernel() {
    curl --unix-socket /tmp/firecracker.socket -i \
      -X PUT 'http://localhost/boot-source'   \
      -H 'Accept: application/json'           \
      -H 'Content-Type: application/json'     \
      -d "{
            \"kernel_image_path\": \"${kernel_path}\",
            \"boot_args\": \"console=ttyS0 reboot=k panic=1 pci=off\"
       }"
}

set_kernel

function set_fs() {
    curl --unix-socket /tmp/firecracker.socket -i \
        -X PUT 'http://localhost/drives/rootfs' \
        -H 'Accept: application/json'           \
        -H 'Content-Type: application/json'     \
        -d "{
            \"drive_id\": \"rootfs\",
            \"path_on_host\": \"${rootfs_path}\",
            \"is_root_device\": true,
            \"is_read_only\": false
        }"
}

set_fs

function set_spec() {
    curl --unix-socket /tmp/firecracker.socket -i  \
        -X PUT 'http://localhost/machine-config' \
        -H 'Accept: application/json'            \
        -H 'Content-Type: application/json'      \
        -d '{
            "vcpu_count": 2,
            "mem_size_mib": 512
        }'
}

set_spec

function set_network() {
    curl --unix-socket /tmp/firecracker.socket -X PUT 'http://localhost/network-interfaces/wlo1' -d '{ "iface_id": "wlo1", "guest_mac": "AA:FC:00:00:00:01", "host_dev_name": "tap0" }'
}

# set_network

# Start VM 
function start_vm() {
    curl --unix-socket /tmp/firecracker.socket -i \
        -X PUT 'http://localhost/actions'       \
        -H  'Accept: application/json'          \
        -H  'Content-Type: application/json'    \
        -d '{
            "action_type": "InstanceStart"
        }'
}

# start_vm

# Stop VM
function stop_vm() {
    curl --unix-socket /tmp/firecracker.socket -i \
        -X PUT 'http://localhost/actions'       \
        -H  'Accept: application/json'          \
        -H  'Content-Type: application/json'    \
        -d '{
            "action_type": "SendCtrlAltDel"
        }'
}

