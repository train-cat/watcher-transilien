package metadata

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Eraac/train-sniffer/utils"
)

type (
	Wave struct {
		UUID        string
		TimeProcess time.Duration
		LaunchedAt  time.Time
	}
)

const indexWave = "wave"

func init() {
	indexes = append(indexes, &Wave{})
}

func (w *Wave) getMappings() (string, string) {
	return indexWave, `{
		"mappings": {
			"wave": {
				"_all": {"enabled": false},
				"properties": {
					"time_process": {"type": "integer"},
					"launched_at":  {"type": "date", "format": "yyyy-MM-dd HH:mm:ss"}
				}
			}
		}
	}`
}

func (w Wave) Persist() {
	_, err := client.Index().
		Index(indexWave).
		Type(indexWave).
		Id(w.UUID).
		BodyJson(w).
		Do(ctx)

	if err != nil {
		utils.Error(err.Error())
	}

	utils.Log(fmt.Sprintf("%+v", w))
}

func (w Wave) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		TimeProcess int64  `json:"time_process"`
		LaunchedAt  string `json:"launched_at"`
	}{
		TimeProcess: w.TimeProcess.Nanoseconds() / 1e6,
		LaunchedAt:  w.LaunchedAt.Format("2006-01-02 15:04:05"),
	})
}
