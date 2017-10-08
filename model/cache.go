package model

import (
	"fmt"
	"time"

	rediscache "github.com/go-redis/cache"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"github.com/train-sh/sniffer-transilien/utils"
	"github.com/vmihailenco/msgpack"
)

const (
	keyTrainCode  = "train-code-%s"
	keyPassage    = "passage-station_id-%d-train_code-%s"
	keyBanStation = "ban-station-%d"
)

type cacheModel struct{}

var (
	ring  *redis.Ring
	codec *rediscache.Codec
	cache cacheModel

	// ErrorMiss is returned when cache is miss
	ErrorMiss = fmt.Errorf("Cache miss")
)

func initCache() {
	ring = redis.NewRing(&redis.RingOptions{
		Addrs: viper.GetStringMapString("redis.servers"),
	})

	codec = &rediscache.Codec{
		Redis: ring,
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}

	if viper.GetBool("redis.flush_all_on_start") {
		err := ring.FlushAll().Err()

		if err != nil {
			utils.Error(err.Error())
		}
	}
}

func (c cacheModel) buildKeyTrainCode(code string) string {
	return fmt.Sprintf(keyTrainCode, code)
}

func (c cacheModel) buildKeyPassage(s Station, code string) string {
	return fmt.Sprintf(keyPassage, s.ID, code)
}

func (c cacheModel) buildKeyBanStation(s Station) string {
	return fmt.Sprintf(keyBanStation, s.ID)
}

func (c cacheModel) get(key string, obj interface{}) error {
	exists, err := ring.Exists(key).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return ErrorMiss
	}

	err = codec.Get(key, &obj)

	if err != nil && err != redis.Nil {
		return err
	}

	return nil
}

func (c cacheModel) set(key string, obj interface{}) error {
	return codec.Set(&rediscache.Item{
		Key:        key,
		Object:     obj,
		Expiration: -1,
	})
}

func (c cacheModel) setBan(key string, h int) error {
	return ring.Set(key, true, time.Hour*time.Duration(h)).Err()
}

func (c cacheModel) IsKeyExist(key string) bool {
	exist, err := ring.Exists(key).Result()

	if err != nil {
		utils.Error(err.Error())
		return false
	}

	return exist > 0
}
