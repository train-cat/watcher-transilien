package model

type Issue struct {
	State       string `json:"state"`
	Code        string `json:"code"`
	Schedule    string `json:"schedule"`
	StationID   uint   `json:"station_id"`
	StationName string `json:"station_name"`
}
