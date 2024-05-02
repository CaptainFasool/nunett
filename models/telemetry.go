package models

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type TelemetryConfig struct {
	ServiceName           string
	OTelCollectorEndpoint string
}

// OpenTelemetryCollector struct definition
type OpenTelemetryCollector struct {
	TracerProvider trace.TracerProvider
	OtEndpoint     string
}

// Collector interface with necessary telemetry functions
type Collector interface {
	Initialize(ctx context.Context) error
	HandleEvent(ctx context.Context, event GEvent) error
	Shutdown(ctx context.Context) error
	GetObservedLevel() ObservabilityLevel
	GetEndpoint() string
}

// Event interface definition
type Event interface {
	ObserveEvent()
}

// EventCategory represents categories of events
type EventCategory int8

// Enumeration of EventCategory
const (
	ACCOUNTING EventCategory = iota + 1
	LOGGING
	TRACING
)

// GEvent represents a generic event implementing the Event interface
type GEvent struct {
	Event
	Observable

	CurrentTimestamp time.Time
	Context          context.Context
	Category         EventCategory
	Message          string
	Collectors       []Collector
}

// Timestamp method returns the current time
func (ge GEvent) Timestamp() time.Time {
	return time.Now()
}

// Observable interface for observability features
type Observable interface {
	GetObservabilityLevel() ObservabilityLevel
	GetCollectors() []Collector
	RegisterCollectors()
}

// ObservabilityLevel defines levels of observability
type ObservabilityLevel float32

// Constants representing levels of observability
const (
	TRACE ObservabilityLevel = 1.0
	DEBUG ObservabilityLevel = 2.0
	INFO  ObservabilityLevel = 3.0
	WARN  ObservabilityLevel = 4.0
	ERROR ObservabilityLevel = 5.0
	FATAL ObservabilityLevel = 6.0
)
