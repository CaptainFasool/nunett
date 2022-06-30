package models

type Metadata struct {
	Name     string `json:"name,omitempty"`
	Resource struct {
		UpdateTimestamp int     `json:"update_timestamp,omitempty"`
		RamMax          int     `json:"ram_max,omitempty"`
		TotalCore       int     `json:"total_core,omitempty"`
		CpuMax          float32 `json:"cpu_max,omitempty"`
		CpuUsage        float32 `json:"cpu_usage,omitempty"`
	} `json:"resource,omitempty"`
	Available struct {
		UpdateTimestamp int `json:"update_timestamp,omitempty"`
		Ram             int `json:"ram,omitempty"`
	} `json:"available,omitempty"`
	Reserved struct {
		Cpu    int `json:"cpu,omitempty"`
		Memory int `json:"memory,omitempty"`
	} `json:"reserved,omitempty"`
	Network   string `json:"network,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
}

type MetadataV2 struct {
	Name            string `json:"name,omitempty"`
	UpdateTimestamp int64  `json:"update_timestamp,omitempty"`
	Resource        struct {
		MemoryMax int64 `json:"memory_max,omitempty"`
		TotalCore int64 `json:"total_core,omitempty"`
		CpuMax    int64 `json:"cpu_max,omitempty"`
	} `json:"resource,omitempty"`
	Available struct {
		CPU    int64 `json:"cpu,omitempty"`
		Memory int64 `json:"memory,omitempty"`
	} `json:"available,omitempty"`
	Reserved struct {
		CPU    int64 `json:"cpu,omitempty"`
		Memory int64 `json:"memory,omitempty"`
	} `json:"reserved,omitempty"`
	Network   string `json:"network,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
}
