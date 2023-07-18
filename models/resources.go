package models

type MB int
type MHz int

// Resources is the base struct used by other resources usage
// related structs. It describes hardware components capacities.
type Resources struct {
	TotCPU  MHz `json:"tot_cpu_mhz"`
	RAM     MB  `json:"ram_mb"`
	VCPU    MHz `json:"vcpu_mhz"`
	Disk    MB  `json:"disk_mb"`
	CoreCPU MHz `json:"cpu_mhz"`
	CPUNo   int `json:"cpu_no"`
}

// FreeResources are the resources free to be used by NuNet,
// subtracting resources being already being used by NuNet
// from onboarded resources.
type FreeResources struct {
	ID uint `json:"id"`
	Resources
}

// AvailableResources describes the machine host's hardware
// resources onboarded to be used on NuNet
type AvailableResources struct {
	ID uint `json:"id"`
	Resources
}

// ServiceResourceRequirements is the resources table used by
// NuNet services
type ServiceResourceRequirements struct {
	ID uint `json:"id"`
	Resources
}
