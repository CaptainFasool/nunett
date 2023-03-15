package machines

// This file replicates the schema of DHTContents for marshaling and unmarshaling
// Currently, the entire schema is divided into 3 parts:
// 1. Machines Index
// 2. Available Resources Index
// 3. Services Index

// 1. Machines Index
type IP []any

type PeerInfo struct {
	NodeID    string `json:"nodeID,omitempty"`
	Key       string `json:"key,omitempty"`
	Mid       string `json:"mid,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	Address   IP     `json:"_address,omitempty"`
}

type AvailableResource struct {
	CpuNo     int     `json:"cpu_no"`
	CpuHz     int     `json:"cpu_hz"`
	PriceCpu  float64 `json:"price_cpu"`
	Ram       int     `json:"ram"`
	PriceRam  float64 `json:"price_ram"`
	Vcpu      int     `json:"vcpu"`
	Disk      float64 `json:"disk"`
	PriceDisk float64 `json:"price_disk"`
}

type GpuInfo struct {
	Name string `json:"name,omitempty"`
	Vram string `json:"vram,omitempty"`
}

type Peer struct {
	PeerInfo             PeerInfo          `json:"peer_info,omitempty"`
	IPAddr               IP                `json:"ip_addr,omitempty"`
	AvailableResources   AvailableResource `json:"available_resources,omitempty"`
	TokenomicsAdress     string            `json:"tokenomics_adress,omitempty"`
	TokenomicsBlockchain string            `json:"tokenomics_blockchain,omitempty"`
	HasGpu               string            `json:"has_gpu,omitempty"`
	AllowCardano         string            `json:"allow_cardano,omitempty"`
	GpuInfo              GpuInfo           `json:"gpu_info,omitempty"`
	Timestamp            uint32            `json:"timestamp,omitempty"`
}

type Machines map[string]Peer

// TODO: 2. Available Resources Index
// TODO: 3. Services Index
