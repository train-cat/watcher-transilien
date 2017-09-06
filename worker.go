package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/Eraac/train-sniffer/metadata"
	"github.com/Eraac/train-sniffer/model"
	"github.com/Eraac/train-sniffer/sncf"
)

type (
	job struct {
		station  model.Station
		passages []sncf.Passage
		UUID     string
	}
)

func consume(jobs chan job) {
	wg := sync.WaitGroup{}

	for sj := range jobs {
		go func(j job) {
			wg.Add(1)
			defer wg.Done()

			md := metadata.Job{StartAt: time.Now(), Station: j.station, WaveID: j.UUID}
			err := j.do()

			if err != nil {
				md.Error = err.Error()
			}

			md.TimeProcess = time.Since(md.StartAt)

			md.Persist()
		}(sj)
	}

	fmt.Println("Queue closed")

	wg.Wait()
}

func (j job) do() error {
	for _, passage := range j.passages {
		// if station has not realtime api available persist information in database
		if passage.Date.Mode == sncf.ModeTheoretical {
			j.station.IsRealTime = false
			err := (&j.station).Update()

			if err != nil {
				return err
			}

			break
		}

		train, err := GetTrain(passage)

		if err != nil {
			return err
		}

		p, err := GetPassage(passage, &j.station, train)

		if err != nil {
			return err
		}

		rt := &model.Realtime{
			State:     GenerateModelState(passage.State),
			PassageID: p.ID,
		}

		schedule, err := passage.GetTime()

		if err != nil {
			return err
		}

		mrt := &metadata.Realtime{
			WaveID: j.UUID,
			CheckedAt: time.Now(),
			State: rt.State,
			Schedule: schedule,
			Station: j.station,
			Train: metadata.Train{Code: passage.TrainID, Mission: passage.Mission},
		}

		err = rt.Persist()

		if err != nil {
			return err
		}

		mrt.Persist()

		// TODO notify if delayed or deleted
	}

	return nil
}
