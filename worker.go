package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/spf13/viper"
	"github.com/train-cat/client-train-go"
	"github.com/train-cat/sniffer-transilien/cache"
	"github.com/train-cat/sniffer-transilien/metadata"
	"github.com/train-cat/sniffer-transilien/model"
	"github.com/train-cat/sniffer-transilien/sncf"
	"github.com/train-cat/sniffer-transilien/utils"
	"google.golang.org/api/option"
)

type (
	job struct {
		station  traincat.Station
		passages []sncf.Passage
		UUID     string
	}
)

var (
	topic *pubsub.Topic

	// ErrorTimeoutPubsub is returned when pubsub not response
	ErrorTimeoutPubsub = fmt.Errorf("pubsub timeout")
)

func init() {
	c, err := pubsub.NewClient(
		context.Background(),
		viper.GetString("pubsub.project"),
		option.WithCredentialsFile(viper.GetString("pubsub.credentials")),
	)

	if err != nil {
		utils.Error(err.Error())
		os.Exit(utils.ErrorInitPubSub)
	}

	topic = c.Topic(viper.GetString("pubsub.topic"))
}

func consume(jobs chan job) {
	wg := sync.WaitGroup{}

	for sj := range jobs {
		go func(j job) {
			wg.Add(1)
			defer wg.Done()

			md := metadata.Job{StartAt: time.Now(), Station: j.station, WaveID: j.UUID}
			err := j.do()

			if err != nil {
				utils.Error(err.Error())
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
			utils.Warning(fmt.Sprintf("stations %s is not realtime", j.station.Name))
			break
		}

		schedule, err := passage.GetTime()

		if err != nil {
			return err
		}

		state := GenerateModelState(passage.State)

		mrt := &metadata.Realtime{
			WaveID:    j.UUID,
			CheckedAt: time.Now(),
			State:     state,
			Schedule:  schedule,
			Station:   j.station,
			Train:     metadata.Train{Code: passage.TrainID, Mission: passage.Mission},
		}

		mrt.Persist()

		if state != StateOnTime {
			err = publish(passage.TrainID, state, j.station, schedule)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func publish(code string, state string, station traincat.Station, schedule time.Time) error {
	i := model.Issue{
		Code:        code,
		State:       state,
		StationID:   station.ID,
		Schedule:    schedule.Format("02/01/2006 15:04 -0700"),
		StationName: station.Name,
	}

	key := cache.BuildKeyIssue(i)

	if cache.IsKeyExist(key) {
		return nil
	}

	cache.SetExpiry(key, 12)

	b, err := json.Marshal(i)

	if err != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result := topic.Publish(ctx, &pubsub.Message{
		Data: b,
	})

	select {
	case <-ctx.Done():
		return ErrorTimeoutPubsub
	case <-result.Ready():
		_, err = result.Get(ctx)
		return err
	}
}
