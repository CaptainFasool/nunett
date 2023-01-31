package models

type AdapterMessage struct {
	Sender    string         `json:"sender"`
	Uid       string         `json:"uid"`
	Timestamp string         `json:"msg_time"`
	Data      GenericMessage `json:"message"`
}

type GenericMessage struct {
	Type    string                 `json:"msg_type"` // See internal/messaging module for options
	Message map[string]interface{} `json:"message"`
}
