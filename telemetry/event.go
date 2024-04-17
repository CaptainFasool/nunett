package telemetry

import time
import dms
import os
import encoding.json


// event interface definition
type Event interface {
	observeEvent()
}

type EventCategory int8

const {
	ACCOUNTING EventCategory = 1
	LOGGING EventCategory = 2
	TRACING EventCategory = 3

}

// generic event type, implemeting the interface
// in struct that can be embedded in other structs
type gEvent struct {
	Event
	Observable

	timestamp Time
	context Context
	category EventCategory
	message string
	collectors []Collector
}

func (ge gEvent) timestamp() Time {
	return time.Now()
}



