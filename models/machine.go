package models

type IP []any

type PeerInfo struct {
	ID        uint   `json:"id"`
	NodeID    string `json:"nodeID,omitempty"`
	Key       string `json:"key,omitempty"`
	Mid       string `json:"mid,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	Address   string `json:"_address,omitempty"`
}

type Machine struct {
	ID                   uint
	NodeId               string
	PeerInfo             int
	IpAddr               string
	AvailableResources   int
	FreeResources        int
	TokenomicsAddress    string
	TokenomicsBlockchain string
}

type FreeResources struct {
	ID        uint    `json:"id"`
	TotCpuHz  int     `json:"tot_cpu_hz"`
	PriceCpu  float64 `json:"price_cpu"`
	Ram       int     `json:"ram"`
	PriceRam  float64 `json:"price_ram"`
	Vcpu      int     `json:"vcpu"`
	Disk      float64 `json:"disk"`
	PriceDisk float64 `json:"price_disk"`
}

type AvailableResources struct {
	ID        uint
	TotCpuHz  int
	CpuNo     int
	CpuHz     float64
	PriceCpu  float64
	Ram       int
	PriceRam  float64
	Vcpu      int
	Disk      float64
	PriceDisk float64
}

type Services struct {
	ID                   uint
	ServiceName          string
	ContainerID          string
	ResourceRequirements int
	ImageID              string
	// TODO: Add ContainerType field

}

type ServiceResourceRequirements struct {
	ID   uint
	CPU  int
	RAM  int
	VCPU int
	HDD  int
}

type Libp2pInfo struct {
	ID         uint   `json:"id"`
	PrivateKey []byte `json:"private_key"`
	PublicKey  []byte `json:"public_key"`
}
