package model

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type (
	Passage struct {
		gorm.Model
		Time      time.Time
		IsWeek    bool
		Station   Station
		StationID uint
		Train     Train
		TrainID   uint
	}

	passageRepository struct{}
)

var PassageRepository *passageRepository

func (r *passageRepository) FindByStationAndTrain(station Station, train Train) (*Passage, error) {
	key := cache.buildKeyPassage(station, train)
	passage, err := cache.getPassage(key)

	if err != nil && err != MissError {
		return nil, err
	}

	if err == nil {
		return passage, nil
	}

	err = db.Where("station_id = ? AND train_id = ?", station.ID, train.ID).Find(passage).Error

	if err == nil {
		cache.set(key, *passage)
	}

	return passage, err
}

func (p *Passage) Persist() error {
	return db.Save(p).Error
}

func (p *Passage) String() string {
	return fmt.Sprintf("[passage] id: %d - station_id: %d - train_id: %d", p.ID, p.StationID, p.TrainID)
}
