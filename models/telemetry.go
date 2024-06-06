package models

import (
	"os"
)

// TelemetryConfig holds the configuration for telemetry.
type TelemetryConfig struct {
	ServiceName        string
	CollectorEndpoint  string
	ObservabilityLevel string
}

// LoadConfigFromEnv loads configuration from environment variables.
func LoadConfigFromEnv() *TelemetryConfig {
	return &TelemetryConfig{
		ServiceName:        os.Getenv("SERVICE_NAME"),
		CollectorEndpoint:  os.Getenv("COLLECTOR_ENDPOINT"),
		ObservabilityLevel: os.Getenv("OBSERVABILITY_LEVEL"),
	}
}

// ObservabilityLevel defines levels of observability.
type ObservabilityLevel string

// Constants representing levels of observability.
const (
	TRACE ObservabilityLevel = "TRACE"
	DEBUG ObservabilityLevel = "DEBUG"
	INFO  ObservabilityLevel = "INFO"
	WARN  ObservabilityLevel = "WARN"
	ERROR ObservabilityLevel = "ERROR"
	FATAL ObservabilityLevel = "FATAL"
)
