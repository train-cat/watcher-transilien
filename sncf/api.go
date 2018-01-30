package sncf

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/spf13/viper"
	"github.com/train-cat/client-train-go"
	"github.com/train-cat/watcher-transilien/utils"
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

func (api *api) GetPassages(s traincat.Station) ([]Passage, error) {
	req, err := http.NewRequest(http.MethodGet, buildURI(s.UIC), nil)

	if err != nil {
		return nil, err
	}

	addHeaders(req)

	resp, err := api.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	v := passages{}

	err = xml.Unmarshal(data, &v)

	if err != nil {
		return nil, err
	}

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
