package telemetry

import (
	"context"
	"fmt"

	"gitlab.com/nunet/device-management-service/models" // Ensure this path is correct
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// NewOpenTelemetryCollector initializes an OpenTelemetryCollector with the specified endpoint.
func NewOpenTelemetryCollector(ctx context.Context, cfg *models.TelemetryConfig) (*models.OpenTelemetryCollector, error) {
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(cfg.OTelCollectorEndpoint),
		otlptracehttp.WithInsecure(), // will change in prod
	)
	exporter, err := otlptracehttp.New(ctx, client)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
		)),
	)

	otel.SetTracerProvider(tracerProvider)

	return &models.OpenTelemetryCollector{
		TracerProvider: tracerProvider,
		OtEndpoint:     cfg.OTelCollectorEndpoint,
	}, nil
}

func (otc *models.OpenTelemetryCollector) Initialize(ctx context.Context) error {
	return nil
}

func (otc *models.OpenTelemetryCollector) HandleEvent(ctx context.Context, event models.GEvent) error {
	if float32(event.GetObservabilityLevel()) >= float32(otc.GetObservedLevel()) {
		tracer := otel.Tracer("OpenTelemetryCollector")
		_, span := tracer.Start(ctx, fmt.Sprintf("%v", event.Category),
			trace.WithAttributes(toOtelAttributes(event.Message)...),
			trace.WithTimestamp(event.Timestamp),
		)
		span.End()
	}
	return nil
}

func (otc *models.OpenTelemetryCollector) Shutdown(ctx context.Context) error {
	if otc.TracerProvider != nil {
		return otc.TracerProvider.Shutdown(ctx)
	}
	return nil
}

func (otc *models.OpenTelemetryCollector) GetObservedLevel() models.ObservabilityLevel {
	return models.TRACE // Observes all levels
}

func (otc *models.OpenTelemetryCollector) GetEndpoint() string {
	return otc.OtEndpoint
}

// Helper function to convert event message to OpenTelemetry attributes
func toOtelAttributes(message string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("message", message),
	}
}
