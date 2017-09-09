package metadata

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Eraac/train-sniffer/model"
	"github.com/Eraac/train-sniffer/utils"
)

type (
	Realtime struct {
		WaveID    string
		CheckedAt time.Time
		State     string
		Schedule  time.Time
		Station   model.Station
		Train     Train
	}
)

const indexRealtime = "realtime"

func init() {
	indexes = append(indexes, &Realtime{})
}

func (r *Realtime) getMappings() (string, string) {
	return indexRealtime, `{
		"mappings": {
			"realtime": {
				"_all": {"enabled": false},
				"properties": {
					"wave_id":    {"type": "keyword"},
					"checked_at": {"type": "date", "format": "yyyy-MM-dd HH:mm:ss Z"},
					"state":      {"type": "keyword"},
					"schedule":   {"type": "date", "format": "yyyy-MM-dd HH:mm Z"},
					"station": {
						"properties": {
							"id":   {"type": "integer"},
							"name": {"type": "keyword"},
							"uic":  {"type": "keyword"}
						}
					},
					"train": {
						"properties": {
							"code":    {"type": "keyword"},
							"mission": {"type": "keyword"}
						}
					}
				}
			}
		}
	}`
}

func (r *Realtime) Persist() {
	_, err := client.Index().
		Index(indexRealtime).
		Type(indexRealtime).
		BodyJson(r).
		Do(ctx)

	if err != nil {
		utils.Error(err.Error())
	}

	utils.Log(fmt.Sprintf("%+v", r))
}

func (r *Realtime) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		WaveID    string  `json:"wave_id"`
		CheckedAt string  `json:"checked_at"`
		State     string  `json:"state"`
		Schedule  string  `json:"schedule"`
		Station   station `json:"station"`
		Train     Train   `json:"train"`
	}{
		WaveID:    r.WaveID,
		CheckedAt: r.CheckedAt.Format("2006-01-02 15:04:05 -0700"),
		State:     r.State,
		Schedule:  r.Schedule.Format("2006-01-02 15:04 -0700"),
		Station:   station{ID: r.Station.ID, Name: r.Station.Name, UIC: r.Station.UIC},
		Train:     Train{Code: r.Train.Code, Mission: r.Train.Mission},
	})
}
