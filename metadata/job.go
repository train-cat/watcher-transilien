package metadata

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/Eraac/train-sniffer/model"
	"github.com/Eraac/train-sniffer/utils"
)

type (
	Job struct {
		WaveID      string
		StartAt     time.Time
		TimeProcess time.Duration
		Station     model.Station
		Error       string
	}
)

const indexJob = "job"

func init() {
	indexes = append(indexes, &Job{})
}

func (j *Job) getMappings() (string, string) {
	return indexJob, `{
		"mappings": {
			"job": {
				"_all": {"enabled": false},
				"properties": {
					"wave_id":      {"type": "keyword"},
					"start_at":     {"type": "date", "format": "yyyy-MM-dd HH:mm:ss Z"},
					"time_process": {"type": "integer"},
					"station": {
						"properties": {
							"id":   {"type": "integer"},
							"name": {"type": "keyword"},
							"uic":  {"type": "keyword"}
						}
					},
					"error": {"type": "text"}
				}
			}
		}
	}`
}

func (j Job) Persist() {
	_, err := client.Index().
		Index(indexJob).
		Type(indexJob).
		BodyJson(j).
		Do(ctx)

	if err != nil {
		utils.Error(err.Error())
	}

	utils.Log(fmt.Sprintf("%+v", j))
}

func (j Job) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		WaveID      string  `json:"wave_id"`
		StartAt     string  `json:"start_at"`
		TimeProcess int64   `json:"time_process"`
		Station     station `json:"station"`
		Error       string  `json:"error"`
	}{
		WaveID:      j.WaveID,
		StartAt:     j.StartAt.Format("2006-01-02 15:04:05 -0700"),
		TimeProcess: j.TimeProcess.Nanoseconds() / 1e6,
		Station:     station{ID: j.Station.ID, Name: j.Station.Name, UIC: j.Station.UIC},
		Error:       j.Error,
	})
}
