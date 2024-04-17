// collector.go
package telemetry

import (
	"context"
	"fmt"
	"telemetry/config"

	// "gitlab.com/nunet/device-management-service/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// NewOpenTelemetryCollector initializes an OpenTelemetryCollector with the specified endpoint.
func NewOpenTelemetryCollector(ctx context.Context, cfg *config.TelemetryConfig) (*OpenTelemetryCollector, error) {
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(cfg.OTelCollectorEndpoint),
		otlptracehttp.WithInsecure(), // we will change this in prod
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

	return &OpenTelemetryCollector{
		TracerProvider: tracerProvider,
		otEndpoint:     cfg.OTelCollectorEndpoint,
	}, nil
}
func (otc *OpenTelemetryCollector) Initialize(ctx context.Context) error {
	// Additional initialization steps can be added here if needed
	return nil
}

func (otc *OpenTelemetryCollector) HandleEvent(ctx context.Context, event gEvent) error {
	if float32(event.GetObservabilityLevel()) >= float32(otc.GetObservedLevel()) {
		tracer := otel.Tracer("OpenTelemetryCollector")
		_, span := tracer.Start(ctx, fmt.Sprintf("%v", event.category),
			trace.WithAttributes(toOtelAttributes(event.message)...),
			trace.WithTimestamp(event.timestamp),
		)
		span.End()
	}
	return nil
}

func (otc *OpenTelemetryCollector) Shutdown(ctx context.Context) error {
	if otc.TracerProvider != nil {
		return otc.TracerProvider.Shutdown(ctx)
	}
	return nil
}

func (otc *OpenTelemetryCollector) GetObservedLevel() ObservabilityLevel {
	return TRACE // As TRACE is the lowest level, it will observe all levels.
}

func (otc *OpenTelemetryCollector) GetEndpoint() string {
	return otc.otEndpoint
}

// Helper function to convert event message to OpenTelemetry attributes
func toOtelAttributes(message string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("message", message),
	}
}
