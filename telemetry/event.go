package telemetry

import time
import dms.config
import os

type Collector interface {
	observe() 
}

type FileCollector struct {
	logFile File	 
}

type OpenTelemetryCollector struct {
	endpoint string
}

type Observable interface {
	timestamp() Time
	context() Context
	message() string
	level() ObservabilityLevel
	observe() bool
}

func timestamp() Time {
	return time.Now()
}

type Message struct {

}

type Event struct {

}

type ObservabilityLevel {
	TRACE = iota
	DEBUG = iota
	INFO = iota
	WARN = iota
	ERROR = iota
	FATAL = iota
}