package model

type (
	// Issue contain all fields to send to PubSub
	Issue struct {
		State string `json:"state"`
		Code  string `json:"code"`
	}
)
