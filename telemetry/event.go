package telemetry

import (
	"log"
	"time"

	"gitlab.com/nunet/device-management-service/models"
)

type EventImpl struct {
	models.GEvent
}

func (e *EventImpl) ObserveEvent() {
	log.Printf("Event observed at %v with message: %s", time.Now(), e.Message)
}
