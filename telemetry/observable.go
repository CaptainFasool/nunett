package telemetry

import (
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
	return models.ObservabilityLevel(o.config.ObservabilityLevel)
}

func (o *ObservableImpl) GetCollectors() []Collector {
	return o.collectors
}

func (o *ObservableImpl) RegisterCollectors(collectors []Collector) {
	o.collectors = collectors
}
