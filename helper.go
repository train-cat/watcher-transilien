package main

import (
	"fmt"
	"time"

	"github.com/Eraac/train-sniffer/sncf"
	"github.com/Eraac/train-sniffer/model"
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

// GetTrain return train from passage and insert in database if not exist
func GetTrain(passage sncf.Passage) (*model.Train, error) {
	key := fmt.Sprintf("train-%s", passage.TrainID)

	lock.Lock(key)
	defer lock.Unlock(key)

	train, err := model.TrainRepository.FindByCode(passage.TrainID)

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// INSERT train if not exist
	if err == gorm.ErrRecordNotFound {
		terminus, err := model.StationRepository.FindByUIC(passage.TerminusID)

		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}

		train = &model.Train{
			Code:       passage.TrainID,
			Mission:    passage.Mission,
			Terminus:   nil,
			TerminusID: nil,
		}

		if terminus != nil {
			train.Terminus = terminus
			train.TerminusID = &terminus.ID
		}

		err = train.Persist()

		if err != nil {
			return nil, err
		}
	}

	return train, nil
}

// GetPassage return model.Passage and insert in database if not exist
func GetPassage(p sncf.Passage, station *model.Station, train *model.Train) (*model.Passage, error) {
	passage, err := model.PassageRepository.FindByStationAndTrain(*station, *train)

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// INSERT passage is not exist
	if err == gorm.ErrRecordNotFound {
		t, _ := p.GetTime()

		passage = &model.Passage{
			Time:    t,
			IsWeek:  IsWeek(t),
			Station: *station,
			Train:   *train,
		}

		err = passage.Persist()

		if err != nil {
			return nil, err
		}
	}

	return passage, nil
}
