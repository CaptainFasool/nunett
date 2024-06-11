package telemetry

import (
	"context"
	"time"

	"gitlab.com/nunet/device-management-service/models"
)

type Collector interface {
	Initialize(ctx context.Context) error
	HandleEvent(ctx context.Context, event GEvent) (string, error)
	Shutdown(ctx context.Context) error
	GetObservedLevel() models.ObservabilityLevel
	GetEndpoint() string
}

// EventCategory represents categories of events.
type EventCategory int8

// Enumeration of EventCategory.
const (
	ACCOUNTING EventCategory = iota + 1
	LOGGING
	TRACING
)

// Event interface defines the structure for an event.
type Event interface {
	ObserveEvent()
}

// GEvent represents a generic event implementing the Event interface.
type GEvent struct {
	Event            Event
	Observable       interface{} // Keeping this as an interface to avoid cyclic dependency
	CurrentTimestamp time.Time
	Context          context.Context
	Category         EventCategory
	Message          string
	Collectors       []Collector
}

// Timestamp method returns the current time.
func (ge GEvent) Timestamp() time.Time {
	return time.Now()
}
