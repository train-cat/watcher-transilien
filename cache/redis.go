package cache

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"github.com/train-cat/client-train-go"
	"github.com/train-cat/watcher-transilien/model"
	"github.com/train-cat/watcher-transilien/utils"
)

const (
	keyBanStation = "ban-station-%d"
	keyIssue      = "issue-station-%d-train-%s"
)

var ring *redis.Ring

// Init redis
func Init() {
	ring = redis.NewRing(&redis.RingOptions{
		Addrs: viper.GetStringMapString("redis.servers"),
	})

	if viper.GetBool("redis.flush_all_on_start") {
		err := ring.FlushAll().Err()

		if err != nil {
			utils.Error(err.Error())
		}
	}
}

// BuildKeyBanStation return key for one station ban
func BuildKeyBanStation(s traincat.Station) string {
	return fmt.Sprintf(keyBanStation, s.ID)
}

// BuildKeyIssue return key for one issue
func BuildKeyIssue(i model.Issue) string {
	return fmt.Sprintf(keyIssue, i.StationID, i.Code)
}

// Ban one station
func Ban(key string, h int) error {
	return SetExpiry(key, h)
}

// SetExpiry create new key/value they can expire
func SetExpiry(key string, h int) error {
	return ring.Set(key, true, time.Hour*time.Duration(h)).Err()
}

// IsKeyExist return true if key currently exist in redis
func IsKeyExist(key string) bool {
	exist, err := ring.Exists(key).Result()

	if err != nil {
		utils.Error(err.Error())
		return false
	}

	return exist > 0
}
