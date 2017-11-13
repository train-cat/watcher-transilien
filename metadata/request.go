package metadata

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/train-cat/client-train-go"
	"github.com/train-cat/sniffer-transilien/utils"
)

type (
	// Request represent all data around one request send to the SNCF API
	Request struct {
		WaveID       string
		ResponseTime time.Duration
		Station      traincat.Station
		CountPassage int
		StatusCode   int
		SendAt       time.Time
		ResponseBody string
		Error        string
	}
)

const indexRequest = "request"

func init() {
	indexes = append(indexes, &Request{})
}

func (r *Request) getMappings() (string, string) {
	return indexRequest, `{
		"mappings": {
			"request": {
				"_all": {"enabled": false},
				"properties": {
					"wave_id":       {"type": "keyword"},
					"response_time": {"type": "integer"},
					"station": {
						"properties": {
							"id":   {"type": "integer"},
							"name": {"type": "keyword"},
							"uic":  {"type": "keyword"}
						}
					},
					"count_passage": {"type": "integer"},
					"status_code":   {"type": "integer"},
					"send_at":       {"type": "date", "format": "yyyy-MM-dd HH:mm:ss Z"},
					"response_body": {"type": "text"},
					"error":         {"type": "text"}
				}
			}
		}
	}`
}

// Persist metadata to ElasticSearch
func (r Request) Persist() {
	_, err := client.Index().
		Index(indexRequest).
		Type(indexRequest).
		BodyJson(r).
		Do(ctx)

	if err != nil {
		utils.Error(err.Error())
	}

	utils.Log(fmt.Sprintf("%+v", r))
}

// MarshalJSON return good formatted json for ElasticSearch
func (r Request) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		WaveID       string  `json:"wave_id"`
		ResponseTime int64   `json:"response_time"`
		Station      station `json:"station"`
		CountPassage int     `json:"count_passage"`
		StatusCode   int     `json:"status_code"`
		SendAt       string  `json:"send_at"`
		ResponseBody string  `json:"response_body"`
		Error        string  `json:"error"`
	}{
		WaveID:       r.WaveID,
		ResponseTime: r.ResponseTime.Nanoseconds() / 1e6,
		Station:      station{ID: r.Station.ID, Name: r.Station.Name, UIC: r.Station.UIC},
		CountPassage: r.CountPassage,
		StatusCode:   r.StatusCode,
		SendAt:       r.SendAt.Format("2006-01-02 15:04:05 -0700"),
		ResponseBody: r.ResponseBody,
		Error:        r.Error,
	})
}
