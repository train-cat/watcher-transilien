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
	"github.com/train-cat/watcher-transilien/cache"
	"github.com/train-cat/watcher-transilien/model"
	"github.com/train-cat/watcher-transilien/sncf"
	"github.com/train-cat/watcher-transilien/utils"
	"google.golang.org/api/option"
)

type (
	job struct {
		station  traincat.Station
		passages []sncf.Passage
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

			if err := j.do(); err != nil {
				utils.Error(err.Error())
			}
		}(sj)
	}

	fmt.Println("Queue closed")

	wg.Wait()
}

func (j job) do() error {
	for _, passage := range j.passages {
		// if station has not realtime ban it
		if passage.Date.Mode == sncf.ModeTheoretical {
			cache.Ban(cache.BuildKeyBanStation(j.station), viper.GetInt("watcher.ban.no_real_time"))
			break
		}

		schedule, err := passage.GetTime()

		if err != nil {
			return err
		}

		state := GenerateModelState(passage.State)

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
