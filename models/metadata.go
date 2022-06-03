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



	// 	"name": "kubuntu",
	// 	"resource": {
	// 	  "update_timestamp": 1651648799,
	// 	  "ram_max": 15843,
	// 	  "total_cores": 8,
	// 	  "cpu_max": 4100.0,
	// 	  "cpu_usage": 5.5
	// 	},
	// 	"available": {
	// 	  "updated_timestamp": 1651648799,
	// 	  "ram": 10695
	// 	},
	// 	"reserved": {
	// 	  "cpu": 27800,
	// 	  "memory": 13843
	// 	},
	// 	"network": "nunet-development",
	// 	"public_key": "0x0541422b9e05e9f0c0c9b393313279aada6eabb2"
	// }
	