package cmd

import (
	"context"
	"fmt"

	"gitlab.com/nunet/device-management-service/config"
	"gitlab.com/nunet/device-management-service/telemetry"
)

func Execute() {
	// Load configuration
	cfg, err := config.LoadConfig("./config.yaml")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	// Setup logging based on the loaded configuration
	setupLogging(cfg.Logging)

	// Initialize telemetry
	collector, err := telemetry.NewOpenTelemetryCollector(context.Background(), &cfg.Telemetry)
	if err != nil {
		fmt.Println("Error initializing telemetry collector:", err)
		return
	}

	// Your other application initialization code here
}

func setupLogging(loggingConfig config.LoggingConfig) {
	// Initialize your logging framework here with the specified level from loggingConfig
	fmt.Printf("Logging level set to: %s\n", loggingConfig.Level)
}
