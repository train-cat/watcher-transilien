package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/train-cat/client-train-go"
	"github.com/train-cat/sniffer-transilien/cache"
	"github.com/train-cat/sniffer-transilien/metadata"
	"github.com/train-cat/sniffer-transilien/sncf"
	"github.com/train-cat/sniffer-transilien/utils"
	"github.com/train-cat/client-train-go/filters"
)

var (
	stations     = []traincat.Station{}
	limitPerWave = 0
	cj           = make(chan job)

	// ErrNoStationsAvailable is returned when no stations is available for realtime
	ErrNoStationsAvailable = fmt.Errorf("no stations available")
)

func init() {
	limitPerWave = viper.GetInt("sniffer.max_call_per_wave")
}

func pull(quit <-chan struct{}) (chan job, error) {
	ticker := time.NewTicker(time.Second * time.Duration(viper.GetInt("sniffer.second_between_call")))
	cs, err := stationDistributor()

	if err != nil {
		return nil, err
	}

	f := wave(cs)

	go func() {
		for {
			select {
			case <-ticker.C:
				f()
			case <-quit:
				fmt.Println("Puller stopped")
				ticker.Stop()
				close(cj)
				return
			}
		}
	}()

	return cj, nil
}

func wave(c <-chan traincat.Station) func() {
	return func() {
		wg := sync.WaitGroup{}
		md := metadata.Wave{UUID: uuid.NewV4().String(), LaunchedAt: time.Now()}

		for i := 0; i < limitPerWave; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				station := <-c
				passages, err := sncf.API.GetPassages(station, md.UUID)

				if err != nil {
					utils.Error(err.Error())
					return // error already logged in metadata.Request
				}

				if len(passages) == 0 {
					cache.Ban(cache.BuildKeyBanStation(station), viper.GetInt("sniffer.ban.no_passage"))
					return
				}

				cj <- job{station: station, passages: passages, UUID: md.UUID}
			}()
		}

		wg.Wait()
		md.TimeProcess = time.Since(md.LaunchedAt)
		md.Persist()
	}
}

func stationDistributor() (<-chan traincat.Station, error) {
	var err error

    max := 100
	f := &filters.Station{Pagination: filters.Pagination{MaxPerPage: &max}}

	stations, err = traincat.CGetAllStations(f)

	if err != nil {
		return nil, err
	}

	l := len(stations)

	if l == 0 {
		return nil, ErrNoStationsAvailable
	}

	c := make(chan traincat.Station)

	go func() {
		// TODO refresh list stations sometimes
		defer close(c)

		i := 0

		for {
			s := stations[i]

			if !IsBan(s) {
				c <- s
			}

			i++

			if i == l {
				i = 0
			}
		}
	}()

	return c, nil
}
