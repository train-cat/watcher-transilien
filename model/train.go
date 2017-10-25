package model

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/train-cat/sniffer-transilien/utils"
)

// Different states of train
const (
	StateOnTime  = "on_time"
	StateDelayed = "delayed"
	StateDeleted = "deleted"
)

type (
	// Train represent one train returned by SNCF API
	Train struct {
		gorm.Model
		Code       string
		Mission    string
		Terminus   *Station
		TerminusID *uint
	}

	trainRepository struct{}
)

// TrainRepository regroup all methods relevant to Train
var TrainRepository *trainRepository

// FindByCode return train by code
func (r *trainRepository) FindByCode(code string) (*Train, error) {
	train := &Train{}

	err := db.Where("code = ?", code).Find(train).Error

	return train, err
}

// IsExist return true if train exist (in redis)
func (r *trainRepository) IsExist(code string) bool {
	key := cache.buildKeyTrainCode(code)

	exist := cache.IsKeyExist(key)

	if exist {
		return exist
	}

	count := 0

	err := db.Model(&Train{}).Where("code = ?", code).Count(&count).Error

	if err != nil {
		utils.Error(err.Error())
		return false
	}

	if count > 0 {
		err = cache.set(key, true)

		if err != nil {
			utils.Error(err.Error())
		}
	}

	return count > 0
}

// Persist train to the database
func (t *Train) Persist() error {
	return db.Save(t).Error
}

func (t *Train) String() string {
	return fmt.Sprintf("[train] id: %d - code: %s - mission: %s - terminus_id: %d - terminus: %v", t.ID, t.Code, t.Mission, t.TerminusID, t.Terminus)
}
