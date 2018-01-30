package main

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/train-cat/client-train-go"
	"github.com/train-cat/watcher-transilien/cache"
	"github.com/train-cat/watcher-transilien/sncf"
	"github.com/train-cat/watcher-transilien/utils"
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
	limitPerWave = viper.GetInt("watcher.max_call_per_wave")
}

func pull(quit <-chan struct{}) (chan job, error) {
	ticker := time.NewTicker(time.Second * time.Duration(viper.GetInt("watcher.second_between_call")))
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
		for i := 0; i < limitPerWave; i++ {
			go func() {

				station := <-c
				passages, err := sncf.API.GetPassages(station)

				if err != nil {
					utils.Error(err.Error())
					return
				}

				if len(passages) == 0 {
					cache.Ban(cache.BuildKeyBanStation(station), viper.GetInt("watcher.ban.no_passage"))
					return
				}

				cj <- job{station: station, passages: passages}
			}()
		}
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
