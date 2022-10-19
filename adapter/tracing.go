package adapter

import (
	"io"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"gitlab.com/nunet/device-management-service/firecracker/telemetry"
)


// newExporter returns a console exporter.
func newExporter(w io.Writer) (tracesdk.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
}

func newResource() *resource.Resource {
	metadata:=telemetry.GetMetadata()
	deviceName:=metadata.Name
	environment:=metadata.Network
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(deviceName),
			attribute.String("environment", environment),
		),
	)
	return r
}

func TracerProvider(url string) (*tracesdk.TracerProvider, error) {
	metadata:=telemetry.GetMetadata()
	environment:=metadata.Network
	deviceName:=metadata.Name
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("Nunet"),
			attribute.String("device name", deviceName),
			attribute.String("environment", environment),
		)),
	)
	return tp, nil
}
