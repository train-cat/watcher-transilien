package model

import (
	"fmt"

	rediscache "github.com/go-redis/cache"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"github.com/vmihailenco/msgpack"
	"gitlab.com/abtasty/rc/utils"
)

const (
	keyTrainCode = "train-code-%s"
	keyPassage   = "passage-station_id-%d-train_id-%d"
)

type cacheModel struct{}

var (
	ring  *redis.Ring
	codec *rediscache.Codec
	cache cacheModel

	MissError        = fmt.Errorf("Cache miss")
	KeyNotFoundError = fmt.Errorf("Key not found")
	NilResponseError = fmt.Errorf("Resource is nil")
)

func init_cache() {
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

func (c cacheModel) buildKeyPassage(s Station, t Train) string {
	return fmt.Sprintf(keyPassage, s.ID, t.ID)
}

func (c cacheModel) getTrain(key string) (*Train, error) {
	train := &Train{}

	err := c.get(key, train)

	return train, err
}

func (c cacheModel) getPassage(key string) (*Passage, error) {
	passage := &Passage{}

	err := c.get(key, passage)

	return passage, err
}

func (c cacheModel) get(key string, obj interface{}) error {
	exists, err := ring.Exists(key).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return MissError
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
