package models

import (
	"time"
)

type DeploymentRequest struct {
	AddressUser string    `json:"address_user"`
	MaxNtx      int       `json:"max_ntx"`
	Blockchain  string    `json:"blockchain"`
	ServiceType string    `json:"service_type"`
	Timestamp   time.Time `json:"timestamp"`
	Params      struct {
		ImageID   string   `json:"image_id"`
		ModelURL  string   `json:"model_url"`
		Packages  []string `json:"packages"`
		NodeID    string   `json:"node_id"`
		PublicKey string   `json:"public_key"`
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
	NodeId  string
	Success bool
	Content string
}