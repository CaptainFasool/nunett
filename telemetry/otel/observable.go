package otel

import (
	"context"
	"log"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ObservableImpl struct {
	collectors []telemetry.Collector
	config     *models.TelemetryConfig
}

func NewObservableImpl(config *models.TelemetryConfig) *ObservableImpl {
	return &ObservableImpl{
		config: config,
	}
}

func (o *ObservableImpl) GetObservabilityLevel() models.ObservabilityLevel {
	return models.ObservabilityLevel(o.config.ObservabilityLevel)
}

func (o *ObservableImpl) GetCollectors() []telemetry.Collector {
	return o.collectors
}

func (o *ObservableImpl) RegisterCollectors(collectors []telemetry.Collector) {
	o.collectors = collectors

	ctx := context.Background()

	// Set up the HTTP trace exporter with the endpoint from the config
	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(o.config.CollectorEndpoint), otlptracehttp.WithInsecure())
	if err != nil {
		log.Printf("Failed to create the HTTP exporter: %v", err)
		return
	}

	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
	otel.SetTracerProvider(tracerProvider)

	collector := &CollectorImpl{
		OpenTelemetryCollector: OpenTelemetryCollector{
			TracerProvider: tracerProvider,
			OtEndpoint:     o.config.CollectorEndpoint,
		},
	}

	o.collectors = append(o.collectors, collector)

	log.Println("Collector registered and initialized")
}
