package telemetry

import (
	"log"
	"strconv"

	"gitlab.com/nunet/device-management-service/models"
)

type Observable interface {
	GetObservabilityLevel() models.ObservabilityLevel
	GetCollectors() []Collector
	RegisterCollectors(collectors []Collector)
}

type ObservableImpl struct {
	collectors []Collector
	config     *models.TelemetryConfig
}

func NewObservableImpl(config *models.TelemetryConfig) *ObservableImpl {
	return &ObservableImpl{
		config: config,
	}
}

func (o *ObservableImpl) GetObservabilityLevel() models.ObservabilityLevel {
	level, err := strconv.Atoi(o.config.ObservabilityLevel)
	if err != nil {
		log.Printf("Invalid observability level: %s, defaulting to INFO", o.config.ObservabilityLevel)
		return models.INFO
	}
	return models.ObservabilityLevel(level)
}

func (o *ObservableImpl) GetCollectors() []Collector {
	return o.collectors
}

func (o *ObservableImpl) RegisterCollectors(collectors []Collector) {
	o.collectors = collectors
}
