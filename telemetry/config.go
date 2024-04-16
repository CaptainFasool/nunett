package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Database  DatabaseConfig  `yaml:"database"`
	Logging   LoggingConfig   `yaml:"logging"`
	Telemetry TelemetryConfig `yaml:"telemetry"`
}

type DatabaseConfig struct {
	Endpoint     string `yaml:"endpoint"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	DatabaseName string `yaml:"databaseName"`
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
