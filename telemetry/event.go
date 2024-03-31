package telemetry

import time
import dms
import os

type Collector interface {
	observeEvent()
	getObservedLevel()
	getEndpoint() 
}

type FileCollector struct {
	logFile string	 
}

type DatabaseCollector struct {
	databaseEndpoint string
}

type OpenTelemetryCollector struct {
	otEndpoint string
}

type ReputationCollector struct {

}

type Observable interface {
	getObservabilityLevel() ObservabilityLevel
	getCollectors() []Collector
}

type Event interface {
	timestamp() Time
	context() Context
	message() string
}

type Message struct {
	// interfaces
	Event
	Observable

	// fields
	sender dms.ID
	receiver dms.ID
	headers Headers
	payload Payload
}

func (m Message) timestamp() Time {
	return time.Now()
}


type LocalEvent struct {
	Event
	Observable
}

// other methods for the 

type ObservabilityLevel float32

const {
	TRACE ObservabilityLevel = 1.0
	DEBUG ObservabilityLevel = 2.0
	INFO ObservabilityLevel = 3.0
	WARN ObservabilityLevel = 4.0
	ERROR ObservabilityLevel = 5.0
	FATAL ObservabilityLevel = 6.0
}