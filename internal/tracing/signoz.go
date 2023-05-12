package tracing

import (
	"context"
	"log"

	"gitlab.com/nunet/device-management-service/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var ServiceName = "NuNet DMS" // TODO: This should be unique to be able to see different DMS in dashboard.

const (
	Insecure = "true"
)

func getAddress() string {
	channelName := utils.GetChannelName()
	var (
		addr string
		// sigNoz Address
		sigNoznunetStagingAddr string = "dev.nunet.io:21002" //"telemetry-staging.nunet.io:14317"
		sigNoznunetTestAddr    string = "dev.nunet.io:21002" //"telemetry-test.nunet.io:4317"
		sigNoznunetEdgeAddr    string = "dev.nunet.io:21002" //"telemetry-edge.nunet.io:34317"
		sigNoznunetTeamAddr    string = "dev.nunet.io:21002" //"telemetry-team.nunet.io:44317"
		sigNozlocalAddr        string = "dev.nunet.io:21002" //"localhost:4317"
	)
	if channelName == "nunet-staging" {
		addr = sigNoznunetStagingAddr
	} else if channelName == "nunet-test" {
		addr = sigNoznunetTestAddr
	} else if channelName == "nunet-edge" {
		addr = sigNoznunetEdgeAddr
	} else if channelName == "nunet-team" {
		addr = sigNoznunetTeamAddr
	} else if channelName == "" { // XXX -- setting empty(not yet onboarded) to test endpoint - not a good idea
		addr = sigNoznunetTestAddr
	} else {
		addr = sigNozlocalAddr
	}

	return addr
}

func InitTracer() func(context.Context) error {

	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if len(Insecure) > 0 {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(getAddress()),
		),
	)

	if err != nil {
		log.Fatal(err)
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", ServiceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Println("Could not set resources: ", err)
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)
	return exporter.Shutdown
}

func Info(msg string, span trace.Span) trace.Span {
	span.AddEvent("{'level':'info','msg':'if response is valid','url':'url','attempt':3,'backoff':'time', 'trace-id':" + span.SpanContext().TraceID().String() + ", 'span-id':" + span.SpanContext().SpanID().String() + "}")
	span.SetAttributes(attribute.String("level", "info"), attribute.String("msg", msg), attribute.Bool("boolean", true))
	return span
}
