package telemetry

type Message struct {
	// a message is a special event
	event gEvent

	// special fields
	sender dms.ID
	receiver dms.ID
	headers Headers
	payload Payload
}
