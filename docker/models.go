package docker

import (
	"time"
)

type LogbinResponse struct {
	ID string `json:"id"`
	RawUrl string `json:"raw_url"`
	CreatedAt time.Time `json:"created_at"`
}

type NewLog struct {
	Title string `json:"title"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

type LogAppend struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}