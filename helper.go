package main

import (
	"fmt"
	"time"

	"github.com/train-sh/sniffer-transilien/sncf"
	"github.com/train-sh/sniffer-transilien/model"
	"github.com/jinzhu/gorm"
	"github.com/xiaonanln/keylock"
)

var (
	lock  = keylock.NewKeyLock()
)

// GenerateModelState return state for database with state from SNCF API
func GenerateModelState(state string) string {
	switch state {
	case sncf.StateTrainDelayed:
		return model.StateDelayed

	case sncf.StateTrainDeleted:
		return model.StateDeleted
	}

	return model.StateOnTime
}

// IsWeek return true if t is on week day
func IsWeek(t time.Time) bool {
	weekday := t.Weekday()

	if weekday == time.Sunday || weekday == time.Saturday {
		return false
	}

	return true
}

// PersistTrain add train to database if not exist
func PersistTrain(passage sncf.Passage) error {
	key := fmt.Sprintf("train-%s", passage.TrainID)

	// Avoid persist train twice if train was found is the same wave
	lock.Lock(key)
	defer lock.Unlock(key)

	exist := model.TrainRepository.IsExist(passage.TrainID)

	if exist {
		return nil
	}

	return createTrain(passage)
}

func createTrain(passage sncf.Passage) error {
	terminus, err := model.StationRepository.FindByUIC(passage.TerminusID)

	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	train := &model.Train{
		Code:       passage.TrainID,
		Mission:    passage.Mission,
		Terminus:   nil,
		TerminusID: nil,
	}

	if terminus != nil {
		train.Terminus = terminus
		train.TerminusID = &terminus.ID
	}

	return train.Persist()
}

// PersistPassage add passage to database if not exist
func PersistPassage(p sncf.Passage, station *model.Station) error {
	exist := model.PassageRepository.IsExist(p.TrainID, station)

	if exist {
		return nil
	}

	return createPassage(p, station)
}

func createPassage(p sncf.Passage, station *model.Station) error {
	t, _ := p.GetTime()

	train, err := model.TrainRepository.FindByCode(p.TrainID)

	if err != nil {
		return err
	}

	passage := &model.Passage{
		Time:    t,
		IsWeek:  IsWeek(t),
		Station: *station,
		Train:   *train,
	}

	return passage.Persist()
}
