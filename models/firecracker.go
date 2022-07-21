package models

// InitVMRequest is the request body for the InitVM endpoint
type InitVMRequest struct {
	SocketFile string `json:"socket_file_path"`
}

type BootSource struct {
	SocketFile string `json:"socket_file_path"`
	KernelPath string `json:"kernel_image_path"`
}

type Drive struct {
	SocketFile  string `json:"socket_file_path"`
	DriveID     string `json:"drive_id"`
	PathOnHost  string `json:"path_on_host"`
	IsRootDrive bool   `json:"is_root_drive"`
	IsReadOnly  bool   `json:"is_read_only"`
}

type MachineConfig struct {
	SocketFile    string `json:"socket_file_path"`
	VCPUCount     int    `json:"vcpu_count"`
	MemorySizeMib int    `json:"mem_size_mib"`
}

type Action struct {
	SocketFile string `json:"socket_file_path"`
	ActionType string `json:"action_type"`
}
