package models

type CapacityForNunet struct {
	Memory         int64  `json:"memory,omitempty"`
	CPU            int64  `json:"cpu,omitempty"`
	Channel        string `json:"channel,omitempty"`
	PaymentAddress string `json:"payment_addr,omitempty"`
	Cardano        bool   `json:"cardano,omitempty"`
}
