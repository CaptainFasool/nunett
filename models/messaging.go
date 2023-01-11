package models

type AdapterMessage struct {
	Sender      string                 `json:"sender"`
	Uid         string                 `json:"uid"`
	Timestamp   string                 `json:"message_time"`
	MessageType string                 `json:"message_type"` // See internal/messaging module for options
	Message     map[string]interface{} `json:"message"`
}
