package model

import (
	"github.com/jinzhu/gorm"
	"fmt"
)

type (
	Train struct {
		gorm.Model
		Code       string
		Mission    string
		Terminus   *Station
		TerminusID *uint
	}

	trainRepository struct{}
)

var TrainRepository *trainRepository

func (r *trainRepository) FindByCode(code string) (*Train, error) {
	key := cache.buildKeyTrainCode(code)
	train, err := cache.getTrain(key)

	if err != nil && err != MissError {
		return nil, err
	}

	if err == nil {
		return train, nil
	}

	err = db.Where("code = ?", code).Find(train).Error

	if err == nil {
		cache.set(key, *train)
	}

	return train, err
}

func (t *Train) Persist() error {
	return db.Save(t).Error
}

func (t *Train) String() string {
	return fmt.Sprintf("[train] id: %d - code: %s - mission: %s - terminus_id: %d - terminus: %v", t.ID, t.Code, t.Mission, t.TerminusID, t.Terminus)
}
