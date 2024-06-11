package otel

import (
	"context"
	"errors"
	"log"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type OpenTelemetryCollector struct {
	TracerProvider *sdktrace.TracerProvider
	OtEndpoint     string
}

type CollectorImpl struct {
	OpenTelemetryCollector
}

func (c *CollectorImpl) Initialize(ctx context.Context) error {
	// Setting up the HTTP trace exporter
	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(c.OtEndpoint))
	if err != nil {
		log.Printf("Failed to create HTTP trace exporter: %v", err)
		return err
	}

	// Setting TracerProvider
	c.TracerProvider = sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
	otel.SetTracerProvider(c.TracerProvider)

	log.Println("Collector initialized with endpoint:", c.OtEndpoint)
	return nil
}

func (c *CollectorImpl) HandleEvent(ctx context.Context, event telemetry.GEvent) (string, error) {
	tr := otel.Tracer("http-tracer")
	ctx, span := tr.Start(ctx, "HandleEvent")
	defer span.End()

	// Add attributes to the span for better trace information
	span.SetAttributes(
		attribute.String("event.message", event.Message),
		attribute.String("event.category", string(event.Category)),
	)

	log.Printf("Handling event: %v", event.Message)

	// Example condition to simulate an error
	if event.Message == "error" {
		err := errors.New("failed to handle event due to XYZ")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	// Mark the span as successful
	span.SetStatus(codes.Ok, "Event processed successfully")
	return "Event processed successfully", nil
}

func (c *CollectorImpl) Shutdown(ctx context.Context) error {
	if err := c.TracerProvider.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down tracer provider: %v", err)
		return err
	}
	log.Println("Collector shutdown successfully")
	return nil
}

func (c *CollectorImpl) GetObservedLevel() models.ObservabilityLevel {
	return models.INFO
}

func (c *CollectorImpl) GetEndpoint() string {
	return c.OtEndpoint
}
