package sncf

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/Eraac/train-sniffer/metadata"
	"github.com/Eraac/train-sniffer/model"
	"github.com/spf13/viper"
)

const (
	ModeTheoretical = "T"
	ModeRealTime    = "R"

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
	API *api
)

func init() {
	API = &api{http.Client{}}
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

func (p Passage) GetTime() (time.Time, error) {
	return time.Parse("02/01/2006 15:04", p.Date.String)
}

func (p Passage) String() string {
	return fmt.Sprintf("[passage] %s (%s) [%s] - %s - %s", p.Date.String, p.Date.Mode, p.State, p.Mission, p.TrainID)
}
