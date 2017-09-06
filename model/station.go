package model

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/jinzhu/gorm"
)

type (
	Station struct {
		gorm.Model
		Name       string
		UIC        string `gorm:"column:UIC"`
		IsRealTime bool   `gorm:"column:is_realtime"`
	}

	stationRepository struct{}

	Stations []*Station
)

var (
	StationRepository *stationRepository

	UpdateError              = fmt.Errorf("Can not update this entity")
	NoStationsAvailableError = fmt.Errorf("No stations available !")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (r *stationRepository) FindByUIC(uic string) (*Station, error) {
	s := &Station{}

	err := db.Where("UIC = ?", uic).Find(s).Error

	if err != nil {
		return nil, err
	}

	return s, nil
}

func (r *stationRepository) FindAllRealtime() (Stations, error) {
	var ss Stations

	err := db.Where("is_realtime = 1").Find(&ss).Error

	if err != nil {
		return nil, err
	}

	return ss, nil
}

func (s *Station) Update() error {
	if db.NewRecord(s) {
		return UpdateError
	}

	return db.Model(s).Update("is_realtime").Error
}

// Ban station from pool for h hours
func (s *Station) Ban(h int) error {
	return cache.setBan(cache.buildKeyBanStation(*s), h)
}

// IsBan return true if station is banned
func (s *Station) IsBan() bool {
	return cache.IsKeyExist(cache.buildKeyBanStation(*s))
}

func (s *Station) String() string {
	return fmt.Sprintf("[station] id: %d - name: %s - UIC: %s", s.ID, s.Name, s.UIC)
}

// StationDistributor loop into slice and through each element into a chan
func StationsDistributor() (<-chan *Station, error) {
	stations, err := StationRepository.FindAllRealtime()

	if err != nil {
		return nil, err
	}

	length := len(stations)

	if length == 0 {
		return nil, NoStationsAvailableError
	}

	c := make(chan *Station)

	go func() {
		// TODO refresh list sometimes
		defer close(c)

		i := 0
		for {
			s := stations[i]

			if !s.IsBan() {
				c <- s
			}

			i++

			if i == length {
				i = 0
			}
		}
	}()

	return c, nil
}
