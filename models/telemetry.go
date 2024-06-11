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
type ObservabilityLevel int

// Constants representing levels of observability.
const (
	TRACE ObservabilityLevel = 1
	DEBUG ObservabilityLevel = 2
	INFO  ObservabilityLevel = 3
	WARN  ObservabilityLevel = 4
	ERROR ObservabilityLevel = 5
	FATAL ObservabilityLevel = 6
)
