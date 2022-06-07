package models

// Provisioned struct holds data about how much total resource
// host machine is equipped with
type Provisioned struct {
	CPU    float64 `json:"cpu,omitempty"`
	Memory uint64  `json:"memory,omitempty"`
}
