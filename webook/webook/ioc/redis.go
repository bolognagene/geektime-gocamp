package ioc

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/config"
	redis "github.com/redis/go-redis/v9"
)

func InitRedis() redis.Cmdable {
	client := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})

	return client
}
