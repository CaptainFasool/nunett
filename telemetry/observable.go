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
	config     *models.TelemetryConfig
}

func NewObservableImpl(config *models.TelemetryConfig) *ObservableImpl {
	return &ObservableImpl{
		config: config, // Initialize with config
	}
}

func (o *ObservableImpl) GetObservabilityLevel() models.ObservabilityLevel {
    switch o.config.ObservabilityLevel {
    case "TRACE":
        return models.TRACE
    case "DEBUG":
        return models.DEBUG
    case "INFO":
        return models.INFO
    case "WARN":
        return models.WARN
    case "ERROR":
        return models.ERROR
    case "FATAL":
        return models.FATAL
    default:
        return models.INFO // Default to INFO if not specified or invalid
    }
}

func (o *ObservableImpl) GetCollectors() []models.Collector {
	return o.collectors
}

func (o *ObservableImpl) RegisterCollectors() {
	// Use the configuration from the struct
	ctx := context.Background()

	// Set up the HTTP trace exporter with the endpoint from the config
	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(o.config.OTelCollectorEndpoint), otlptracehttp.WithInsecure())
	if err != nil {
		log.Printf("Failed to create the HTTP exporter: %v", err)
		return
	}

	tracerProvider := trace.NewTracerProvider(trace.WithBatcher(exp))
	otel.SetTracerProvider(tracerProvider)

	collector := &CollectorImpl{
		models.OpenTelemetryCollector{
			TracerProvider: tracerProvider,
			OtEndpoint:     o.config.OTelCollectorEndpoint,
		},
	}

	o.collectors = append(o.collectors, collector)

	log.Println("Collector registered and initialized")
}
