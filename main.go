package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"gitlab.com/nunet/device-management-service/telemetry"
)

func main() {
	// Load configuration
	cfg, err := telemetry.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Setup logging based on the loaded configuration
	setupLogging(cfg.Logging)

	// Initialize telemetry
	collector, err := telemetry.NewOpenTelemetryCollector(context.Background(), &cfg.Telemetry)
	if err != nil {
		log.Fatalf("Error initializing telemetry collector: %v", err)
	}

	// Setup HTTP server and routes
	setupServer(collector)
}

func setupLogging(loggingConfig config.LoggingConfig) {
	fmt.Printf("Logging level set to: %s\n", loggingConfig.Level)
}

func setupServer(collector *telemetry.OpenTelemetryCollector) {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		err := collector.HandleEvent(ctx, telemetry.gEvent{
			Category:  telemetry.TRACING,
			Message:   "Test event triggered",
			Timestamp: time.Now(),
		})
		if err != nil {
			http.Error(w, "Error handling event", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Event processed successfully"))
	})

	fmt.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
