package models

// NetworkSpec is a stub. Please expand based on requirements.
type NetworkSpec struct {
}

// NetConfig is a stub. Please expand it or completely change it based on requirements.
type NetConfig struct {
	NetworkSpec NetworkSpec `json:"network_spec"` // Network specification
}

// NetStat is a stub. Please expand it or completely change it based on requirements.
type NetStat struct {
	Status string `json:"status"` // Network status
	Info   string `json:"info"`   // Network information
}

// MessageInfo is a stub. Please expand it or completely change it based on requirements.
type MessageInfo struct {
	Info string `json:"info"` // Message information
}

// NetP2P is a stub.
type NetP2P struct {
}
