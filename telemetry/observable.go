package telemetry

import time
import dms
import os
import encoding.json


type Observable interface {
	getObservabilityLevel() ObservabilityLevel
	getCollectors() []Collector
	registerCollectors()
}

type ObservabilityLevel float32

const {
	TRACE ObservabilityLevel = 1.0
	DEBUG ObservabilityLevel = 2.0
	INFO ObservabilityLevel = 3.0
	WARN ObservabilityLevel = 4.0
	ERROR ObservabilityLevel = 5.0
	FATAL ObservabilityLevel = 6.0
}
