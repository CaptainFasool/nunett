package models

// LaunchVMRequest is the request body for the LaunchVM endpoint
type LaunchVMRequest struct {
	SocketFile string `json:"socket_file"`
	ConfigFile string `json:"config_file"`
}
