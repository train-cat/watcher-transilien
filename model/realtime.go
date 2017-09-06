package model

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

const (
	StateOnTime  = "on_time"
	StateDelayed = "delayed"
	StateDeleted = "deleted"
)

type (
	Realtime struct {
		gorm.Model
		State     string
		Passage   Passage
		PassageID uint
	}

	realtimeRepository struct{}
)

var RealtimeRepository *realtimeRepository

func (r *Realtime) Persist() error {
	return db.Save(r).Error
}

func (r *Realtime) String() string {
	return fmt.Sprintf("[realtime] id: %d - State: %s - passage_id: %d", r.ID, r.State, r.PassageID)
}
