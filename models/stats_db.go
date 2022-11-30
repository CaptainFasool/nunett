package models

// NewDeviceOnboarded defines the schema of the data to be sent to stats db when a new device gets onboarded
type NewDeviceOnboarded struct {
	PeerID        string
	CPU           float32
	RAM           float32
	Network       float32
	DedicatedTime float32
	Timestamp     float32
}

// DeviceStatusChange defines the schema of the data to be sent to stats db when a device status gets changed
type DeviceStatusChange struct {
	PeerID    string
	Status    string
	Timestamp float32
}

// DeviceResourceChange defines the schema of the data to be sent to stats db when a device resource gets changed
type DeviceResourceChange struct {
	PeerID                   string
	ChangedAttributeAndValue struct {
		CPU           float32
		RAM           float32
		Network       float32
		DedicatedTime float32
	}
	Timestamp float32
}

// DeviceResourceConfig defines the schema of the data to be sent to stats db when a device resource config gets changed
type DeviceResourceConfig struct {
	PeerID                   string
	ChangedAttributeAndValue struct {
		CPU           float32
		RAM           float32
		Network       float32
		DedicatedTime float32
	}
	Timestamp float32
}

// NewService defines the schema of the data to be sent to stats db when a new service gets registered in the platform
type NewService struct {
	ServiceID          string
	ServiceName        string
	ServiceDescription string
	Timestamp          float32
}

// ServiceStatus defines the schema of the data to be sent to stats db when an existing service gets offline or not operational
type ServiceStatus struct {
	ServiceID           string
	PeerID              string
	OnlineOfflineStatus string
	Timestamp           float32
}

// ServiceCall defines the schema of the data to be sent to stats db when a host machine accepts a deployement request
type ServiceCall struct {
	CallID              float32
	PeerIDOfServiceHost string
	ServiceID           string
	CPUUsed             float32
	MaxRAM              float32
	MemoryUsed          float32
	NetworkBwUsed       float32
	TimeTaken           float32
	Status              string
	Timestamp           float32
}

// ServiceRun defines the schema of the data to be sent to stats db about the status of the job being executed on host machine
type ServiceRun struct {
	CallID              float32
	PeerIDOfServiceHost string
	Status              string
	Timestamp           float32
}

// ServiceRemove defines the schema of the data to be sent to stats db when a new service gets removed from the platform
type ServiceRemove struct {
	ServiceID string
	Timestamp float32
}

// NtxPayment defines the schema of the data to be sent to stats db when a payment is made to device for the completion of service.
type NtxPayment struct {
	ServiceID   string
	AmountOfNtx float32
	PeerID      string
	Timestamp   float32
}