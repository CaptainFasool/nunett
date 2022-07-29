package models

// AddressPrivKey holds Ethereum wallet address and private key from which the
// address is derived.
type AddressPrivKey struct {
	Address    string `json:"address,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
}

// CapacityForNunet is a struct required in request body for the onboarding
type CapacityForNunet struct {
	Memory         int64  `json:"memory,omitempty"`
	CPU            int64  `json:"cpu,omitempty"`
	Channel        string `json:"channel,omitempty"`
	PaymentAddress string `json:"payment_addr,omitempty"`
	Cardano        bool   `json:"cardano,omitempty"`
}

// Provisioned struct holds data about how much total resource
// host machine is equipped with
type Provisioned struct {
	CPU      float64 `json:"cpu,omitempty"`
	Memory   uint64  `json:"memory,omitempty"`
	NumCores uint64  `json:"total_cores,omitempty"`
}

// Metadata has an older version of schema for metadata.json.
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

// MetadataV2 has a newer version of schema for metadata.json.
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