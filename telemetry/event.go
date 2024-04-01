package telemetry

import time
import dms
import os
import encoding.json


// event interface definition
type Event interface {
	timestamp() Time
	context() Context
	message() string
}

// generic event type, implemeting the interface
// in struct that can be embedded in other structs
type gEvent struct {
	Event
	Observable
}

func (ge gEvent) timestamp() Time {
	return time.Now()
}


