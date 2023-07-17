package models

type MB int
type GB int
type MHz int

type Resources struct {
	TotCPU  MHz `json:"tot_cpu_mhz"`
	RAM     MB  `json:"ram_mb"`
	VCPU    MHz `json:"vcpu_mb"`
	Disk    MB  `json:"disk_mb"`
	CoreCPU MHz `json:"cpu_mhz"`
	CPUNo   int `json:"cpu_no"`
}

type FreeResources struct {
	ID uint `json:"id"`
	Resources
}

type AvailableResources struct {
	ID uint `json:"id"`
	Resources
}

type ServiceResourceRequirements struct {
	ID uint `json:"id"`
	Resources
}
