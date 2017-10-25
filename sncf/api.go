package sncf

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/spf13/viper"
	"github.com/train-cat/sniffer-transilien/metadata"
	"github.com/train-cat/sniffer-transilien/model"
	"github.com/train-cat/sniffer-transilien/utils"
)

// Different mode returned by SNCF API for a station
const (
	ModeTheoretical = "T"
	ModeRealTime    = "R"
)

// Different state returned by SNCF API for one passage
const (
	StateTrainDelayed = "Retardé"
	StateTrainDeleted = "Supprimé"
)

type (
	api struct {
		client http.Client
	}

	passages struct {
		XMLName  xml.Name  `xml:"passages"`
		Passages []Passage `xml:"train"`
	}

	// Passage represent response from SNCF API
	Passage struct {
		XMLName xml.Name `xml:"train"`
		Date    struct {
			XMLName xml.Name `xml:"date"`
			Mode    string   `xml:"mode,attr"`
			String  string   `xml:",innerxml"`
		}
		TrainID    string `xml:"num"`
		Mission    string `xml:"miss"`
		TerminusID string `xml:"term"`
		State      string `xml:"etat"`
	}
)

var (
	// API is client for request to SNCF API
	API *api

	location *time.Location
)

// Init client for SNCF API
func Init() {
	API = &api{http.Client{}}

	l, err := time.LoadLocation(viper.GetString("sncf_api.timezone"))

	if err != nil {
		utils.Error(err.Error())
		os.Exit(utils.ErrorLoadTimezone)
	}

	utils.Log(fmt.Sprintf("timezone: %s", l.String()))

	location = l
}

func addHeaders(r *http.Request) {
	r.Header.Add("Content-Type", viper.GetString("sncf_api.headers.content_type"))
	r.Header.Add("Accept", viper.GetString("sncf_api.headers.accept"))
	r.SetBasicAuth(
		viper.GetString("sncf_api.username"),
		viper.GetString("sncf_api.password"),
	)
}

func (api *api) GetPassages(s *model.Station, waveID string) ([]Passage, error) {
	metadata := metadata.Request{Station: *s, WaveID: waveID}

	req, err := http.NewRequest(http.MethodGet, buildURI(s.UIC), nil)

	if err != nil {
		metadata.Error = err.Error()
		metadata.Persist()
		return nil, err
	}

	addHeaders(req)

	metadata.SendAt = time.Now()
	resp, err := api.client.Do(req)
	metadata.ResponseTime = time.Since(metadata.SendAt)
	metadata.StatusCode = resp.StatusCode

	if err != nil {
		metadata.Error = err.Error()
		metadata.Persist()
		return nil, err
	}
	defer resp.Body.Close()

	if metadata.StatusCode != http.StatusOK {
		bs, _ := httputil.DumpResponse(resp, true)
		metadata.ResponseBody = string(bs[:])
		metadata.Persist()
		return nil, fmt.Errorf("%s", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		metadata.Error = err.Error()
		metadata.Persist()
		return nil, err
	}

	v := passages{}

	err = xml.Unmarshal(data, &v)

	if err != nil {
		metadata.Error = err.Error()
		metadata.Persist()
		return nil, err
	}

	metadata.CountPassage = len(v.Passages)
	metadata.Persist()

	return v.Passages, nil
}

func buildURI(stationID string) string {
	return fmt.Sprintf(
		"%s/gare/%s/depart/",
		viper.GetString("sncf_api.host"),
		stationID,
	)
}

// GetTime parse time from SNCF API to time.Time
func (p Passage) GetTime() (time.Time, error) {
	return time.ParseInLocation("02/01/2006 15:04", p.Date.String, location)
}

func (p Passage) String() string {
	return fmt.Sprintf("[passage] %s (%s) [%s] - %s - %s", p.Date.String, p.Date.Mode, p.State, p.Mission, p.TrainID)
}
