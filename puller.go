package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/Eraac/train-sniffer/metadata"
	"github.com/Eraac/train-sniffer/model"
	"github.com/Eraac/train-sniffer/sncf"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"github.com/Eraac/train-sniffer/utils"
)

var (
	limitPerWave = 0
	cj           = make(chan job)
)

func init() {
	limitPerWave = viper.GetInt("sniffer.max_call_per_wave")
}

func pull(quit <-chan struct{}) (chan job, error) {
	ticker := time.NewTicker(time.Second * time.Duration(viper.GetInt("sniffer.second_between_call")))
	cs, err := model.StationsDistributor()

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

func wave(c <-chan *model.Station) func() {
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
					station.Ban(viper.GetInt("sniffer.ban.no_passage"))
					return
				}

				cj <- job{station: *station, passages: passages, UUID: md.UUID}
			}()
		}

		wg.Wait()
		md.TimeProcess = time.Since(md.LaunchedAt)
		md.Persist()
	}
}
