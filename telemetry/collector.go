package telemetry

import (
	"context"
	"fmt"
	"log"

	"gitlab.com/nunet/device-management-service/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

type CollectorImpl struct {
	models.OpenTelemetryCollector
}

func (c *CollectorImpl) Initialize(ctx context.Context) error {
	// Setting up the HTTP trace exporter
	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(c.OtEndpoint))
	if err != nil {
		return err
	}

	// Setting TracerProvider
	c.TracerProvider = trace.NewTracerProvider(trace.WithBatcher(exp))
	otel.SetTracerProvider(c.TracerProvider)

	log.Println("Collector initialized")
	return nil
}

func (c *CollectorImpl) HandleEvent(ctx context.Context, event models.GEvent) error {
	tr := otel.Tracer("http-tracer")
	_, span := tr.Start(ctx, "HandleEvent")
	defer span.End()

	log.Printf("Handling event: %v", event.Message)
	return nil
}

func (c *CollectorImpl) Shutdown(ctx context.Context) error {
	sdkTP, ok := c.TracerProvider.(*trace.TracerProvider)
	if !ok {
		return fmt.Errorf("incorrect tracer provider type")
	}
	if err := sdkTP.Shutdown(ctx); err != nil {
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
