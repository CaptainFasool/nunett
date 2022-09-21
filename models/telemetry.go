package models

type FreeResources struct {
	CPU    float64 `json:"cpu,omitempty"`
	Memory uint64  `json:"memory,omitempty"`
}
