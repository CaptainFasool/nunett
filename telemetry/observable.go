package telemetry

import (
	"context"
	"log"

	"gitlab.com/nunet/device-management-service/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

type ObservableImpl struct {
	collectors []models.Collector // Adding the collectors slice here
}

func (o *ObservableImpl) GetObservabilityLevel() models.ObservabilityLevel {
	return models.INFO // Defaulting to INFO level
}

func (o *ObservableImpl) GetCollectors() []models.Collector {
	return o.collectors
}

func (o *ObservableImpl) RegisterCollectors() {
	config := models.TelemetryConfig{
		ServiceName:           "DemoNunetGoService",
		OTelCollectorEndpoint: "95.216.182.94:4318",
	}

	// Create a context with background for initialization
	ctx := context.Background()

	// Setting up the HTTP trace exporter with the endpoint from the config
	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(config.OTelCollectorEndpoint))
	if err != nil {
		log.Printf("Failed to create exporter: %v", err)
		return
	}

	// Setting TracerProvider
	tracerProvider := trace.NewTracerProvider(trace.WithBatcher(exp))
	otel.SetTracerProvider(tracerProvider)

	// Creating a new collector instance with the configured tracer provider
	collector := &CollectorImpl{
		models.OpenTelemetryCollector{
			TracerProvider: tracerProvider,
			OtEndpoint:     config.OTelCollectorEndpoint,
		},
	}

	// Add the new collector to the list
	o.collectors = append(o.collectors, collector)

	log.Println("Collector registered and initialized")
}
