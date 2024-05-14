package main

//	@title			Device Management Service
//	@version		0.4.169
//	@description	A dashboard application for computing providers.
//	@termsOfService	https://nunet.io/tos

//	@contact.name	Support
//	@contact.url	https://devexchange.nunet.io/
//	@contact.email	support@nunet.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @host		localhost:9999
// @BasePath	/api/v1

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/joho/godotenv"
	"gitlab.com/nunet/device-management-service/cmd"
	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/telemetry"
	"go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	// Define and parse the command-line flag
	testTelemetry := flag.Bool("testtelemetry", false, "Enable telemetry simulation")
	flag.Parse()

	if *testTelemetry {
		// Load .env file
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}

		// Load configuration from environment variables
		config := models.LoadConfigFromEnv()
		if config.ServiceName == "" || config.OTelCollectorEndpoint == "" {
			log.Fatal("Configuration error: Make sure all environment variables are set.")
		}

		// Create an Observable instance with the loaded configuration
		observable := telemetry.NewObservableImpl(config)

		// Register collectors
		observable.RegisterCollectors()

		// Run simulation
		simulateTelemetry(observable)

		// Ensure that the flushing of traces is done before the application exits
		defer forceFlushTracers(observable)
	}

	// Execute command-line interface; should be the last call in main()
	cmd.Execute()
}

// simulateTelemetry generates test telemetry data
func simulateTelemetry(observable *telemetry.ObservableImpl) {
	// Create a context with background for the operation
	ctx := context.Background()

	// Creating a test event
	event := models.GEvent{
		CurrentTimestamp: time.Now(),
		Context:          ctx,
		Category:         models.TRACING,
		Message:          "Test Event: demo trace from dms with env and flag",
		Collectors:       observable.GetCollectors(),
	}

	log.Printf("Simulating event handling for event: %s", event.Message)

	// Handle event using each registered collector and log responses
	for _, collector := range event.Collectors {
		if response, err := collector.HandleEvent(ctx, event); err != nil {
			log.Printf("Error handling event: %v", err)
		} else {
			log.Printf("Event handled successfully: %s, response: %s", event.Message, response)
		}
	}
}

// forceFlushTracers forces all registered collectors to flush their traces
func forceFlushTracers(observable *telemetry.ObservableImpl) {
	for _, collectorInterface := range observable.GetCollectors() {
		if collector, ok := collectorInterface.(*telemetry.CollectorImpl); ok {
			if tp, ok := collector.TracerProvider.(*trace.TracerProvider); ok {
				if err := tp.ForceFlush(context.Background()); err != nil {
					log.Printf("Failed to force flush traces: %v", err)
				} else {
					log.Println("Traces successfully flushed")
				}
			}
		}
	}
}
