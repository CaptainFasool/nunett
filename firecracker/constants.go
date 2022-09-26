package firecracker

const (
	DMS_BASE_URL              = "http://localhost:9999/api/v1"
	ERR_DRIVES_REQ            = "Error in making PUT request to /drives with give body"
	ERR_MACHINE_CONFIG_REQ    = "Error in making PUT request to /machine-config with give body"
	ERR_NETWORK_INTERFACE_REQ = "Error in making PUT request to /network-interfaces with give body"
	ERR_ACTIONS_REQ              = "Error in making PUT request to /actions with give body"
	ERR_BOOTSOURCE_REQ           = "Error in making PUT request to /boot-source with give body"
	FIRECRACKER_KERNEL           = "https://gitlab.com/deependra.singh2/file_img/-/raw/main/hello-vmlinux.bin"
	FIRECRACKER_KERNEL_CHECKSUM  = "http://localhost:8000/sum"
	FIRECRACKER_KERNEL_LOCATION  = "/etc/nunet/cardano/"
)
