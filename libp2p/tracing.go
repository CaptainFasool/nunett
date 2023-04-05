package libp2p

import (
	"encoding/json"
	"io"
	"net/http"

	"gitlab.com/nunet/device-management-service/models"
	"gitlab.com/nunet/device-management-service/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func GetMetadata() models.Metadata {
	resp, err := http.Get(utils.DMSBaseURL + "/onboarding/metadata")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var metadata models.Metadata

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &metadata)
	if err != nil && resp.StatusCode == 200 {
		panic(err)
	}
	zlog.Sugar().Infof("metadata.Reserved: %v", metadata.Reserved)
	return metadata
}

func TracerProvider(url string) (*tracesdk.TracerProvider, error) {
	metadata := GetMetadata()
	environment := metadata.Network
	deviceName := metadata.Name
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
