package models

type BootSource struct {
	KernelImagePath string `json:"kernel_image_path"`
	BootArgs        string `json:"boot_args"`
}

type Drives struct {
	DriveID      string `json:"drive_id"`
	PathOnHost   string `json:"path_on_host"`
	IsRootDevice bool   `json:"is_root_device"`
	IsReadOnly   bool   `json:"is_read_only"`
}

type MachineConfig struct {
	VCPUCount  int `json:"vcpu_count"`
	MemSizeMib int `json:"mem_size_mib"`
}

type NetworkInterfaces struct {
	IfaceID     string `json:"iface_id"`
	GuestMac    string `json:"guest_mac"`
	HostDevName string `json:"host_dev_name"`
}

type Actions struct {
	ActionType string `json:"action_type"`
}

type VirtualMachine struct {
	ID               uint   `json:"id"`
	SocketFile       string `json:"socket_file"`
	BootSource       string `json:"boot_source"`
	Filesystem       string `json:"filesystem"`
	NetworkInterface string `json:"network_interface"`
	State            string `json:"state"`
}
