package cache

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"github.com/train-cat/client-train-go"
	"github.com/train-cat/sniffer-transilien/model"
	"github.com/train-cat/sniffer-transilien/utils"
)

const (
	keyTrainCode  = "train-code-%s"
	keyPassage    = "passage-station_id-%d-train_code-%s"
	keyBanStation = "ban-station-%d"
	keyIssue      = "issue-station-%d-train-%s"
)

var ring *redis.Ring

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

func BuildKeyTrainCode(code string) string {
	return fmt.Sprintf(keyTrainCode, code)
}

func BuildKeyPassage(s traincat.Station, code string) string {
	return fmt.Sprintf(keyPassage, s.ID, code)
}

func BuildKeyBanStation(s traincat.Station) string {
	return fmt.Sprintf(keyBanStation, s.ID)
}

func BuildKeyIssue(i model.Issue) string {
	return fmt.Sprintf(keyIssue, i.StationID, i.Code)
}

func Set(key string, obj interface{}) error {
	return ring.Set(key, obj, -1).Err()
}

func Ban(key string, h int) error {
	return SetExpiry(key, h)
}

func SetExpiry(key string, h int) error {
	return ring.Set(key, true, time.Hour*time.Duration(h)).Err()
}

func IsKeyExist(key string) bool {
	exist, err := ring.Exists(key).Result()

	if err != nil {
		utils.Error(err.Error())
		return false
	}

	return exist > 0
}
