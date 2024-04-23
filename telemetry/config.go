package telemetry

import (
	"os"

	"gitlab.com/nunet/device-management-service/models"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Logging   LoggingConfig   `yaml:"logging"`
	Telemetry TelemetryConfig `yaml:"telemetry"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type TelemetryConfig struct {
	ServiceName           string `yaml:"service_name"`
	OTelCollectorEndpoint string `yaml:"otel_collector_endpoint"`
}

func LoadConfig(filePath string) (*Config, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err = decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

type collectorInit models.OpenTelemetryCollector
