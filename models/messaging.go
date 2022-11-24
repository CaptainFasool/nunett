package models

type AdapterMessage struct {
	Sender    string `json:"sender"`
	Uid       string `json:"uid"`
	Timestamp string `json:"msg_time"`
	Message   string `json:"message"`
}
