package main

import (
	"time"

	"github.com/train-cat/client-train-go"
	"github.com/train-cat/sniffer-transilien/cache"
	"github.com/train-cat/sniffer-transilien/sncf"
	"github.com/xiaonanln/keylock"
)

var (
	lock = keylock.NewKeyLock()
)

const (
	StateOnTime  = "on_time"
	StateDelayed = "delayed"
	StateDeleted = "deleted"
)

// GenerateModelState return state for database with state from SNCF API
func GenerateModelState(state string) string {
	switch state {
	case sncf.StateTrainDelayed:
		return StateDelayed

	case sncf.StateTrainDeleted:
		return StateDeleted
	}

	return StateOnTime
}

// IsWeek return true if t is on week day
func IsWeek(t time.Time) bool {
	weekday := t.Weekday()

	if weekday == time.Sunday || weekday == time.Saturday {
		return false
	}

	return true
}

func IsBan(s traincat.Station) bool {
	return cache.IsKeyExist(cache.BuildKeyBanStation(s))
}

func terminusIDByUIC(uic string) uint {
	for _, s := range stations {
		if s.UIC == uic {
			return s.ID
		}
	}

	return 0
}

// PersistTrain add train to database if not exist
func PersistTrain(passage sncf.Passage) error {
	key := cache.BuildKeyTrainCode(passage.TrainID)

	// Avoid persist train twice if train was found is the same wave
	lock.Lock(key)
	defer lock.Unlock(key)

	if cache.IsKeyExist(key) {
		return nil
	}

	exist, err := traincat.TrainExist(passage.TrainID)

	if exist {
		return cache.Set(key, true)
	}

	if err != nil {
		return err
	}

	return createTrain(passage)
}

func createTrain(passage sncf.Passage) error {
	terminus := terminusIDByUIC(passage.TerminusID)

	train := traincat.TrainInput{
		Code:    &passage.TrainID,
		Mission: &passage.Mission,
	}

	if terminus != 0 {
		train.TerminusID = &terminus
	}

	_, err := traincat.PostTrain(train)

	return err
}

// PersistPassage add passage to database if not exist
func PersistPassage(p sncf.Passage, station traincat.Station) error {
	key := cache.BuildKeyPassage(station, p.TrainID)

	if cache.IsKeyExist(key) {
		return nil
	}

	exist, err := traincat.StopExist(station.ID, p.TrainID)

	if exist {
		return cache.Set(key, true)
	}

	if err != nil {
		return err
	}

	return createPassage(p, station)
}

func createPassage(p sncf.Passage, station traincat.Station) error {
	t, err := p.GetTime()

	if err != nil {
		return err
	}

	stop := traincat.StopInput{
		Schedule: t.Format("15:04"),
		IsWeek:   IsWeek(t),
	}

	_, err = traincat.PostStop(station.ID, p.TrainID, stop)

	return err
}
