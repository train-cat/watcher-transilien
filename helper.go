package main

import (
	"github.com/train-cat/client-train-go"
	"github.com/train-cat/watcher-transilien/cache"
	"github.com/train-cat/watcher-transilien/sncf"
	"github.com/xiaonanln/keylock"
)

var (
	lock = keylock.NewKeyLock()
)

// All state available from API (in english)
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

// IsBan return true if station is currently ban from the puller
func IsBan(s traincat.Station) bool {
	return cache.IsKeyExist(cache.BuildKeyBanStation(s))
}
