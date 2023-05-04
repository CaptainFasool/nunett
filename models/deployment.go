package models

import (
	"time"

	"gorm.io/gorm"
)

type DeploymentRequest struct {
	RequesterWalletAddress string    `json:"address_user"` // service provider wallet address
	MaxNtx                 int       `json:"max_ntx"`
	Blockchain             string    `json:"blockchain"`
	ServiceType            string    `json:"service_type"`
	Timestamp              time.Time `json:"timestamp"`
	Params                 struct {
		ImageID     string   `json:"image_id"`
		ModelURL    string   `json:"model_url"`
		Packages    []string `json:"packages"`
		RemoteNodeID      string   `json:"node_id"` // NodeID of compute provider (machine to deploy the job on)
		RemotePublicKey   string   `json:"public_key"` // Public key of compute provider
		LocalNodeID      string   `json:"local_node_id"` // NodeID of service provider (machine triggering the job)
		LocalPublicKey   string   `json:"local_public_key"` // Public key of service provider
		MachineType string   `json:"machine_type"`
	} `json:"params"`
	Constraints struct {
		Complexity string `json:"complexity"`
		CPU        int    `json:"cpu"`
		RAM        int    `json:"ram"`
		Vram       int    `json:"vram"`
		Power      int    `json:"power"`
		Time       int    `json:"time"`
	} `json:"constraints"`
	TraceInfo struct {
		TraceID     string `json:"trace_id"`
		SpanID      string `json:"span_id"`
		TraceFlags  string `json:"trace_flags"`
		TraceStates string `json:"trace_state"`
	} `json:"traceinfo"`
}

type DeploymentResponse struct {
	Success bool   `json:"success"`
	Content string `json:"content"`
}

type DeploymentRequestFlat struct {
	gorm.Model
	DeploymentRequest string `json:"deployment_request"`
}
