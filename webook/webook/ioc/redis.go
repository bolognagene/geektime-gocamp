package ioc

import (
	redis "github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() redis.Cmdable {
	// 不用viper的做法
	/*client := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})*/

	// 用viper的做法
	addr := viper.GetString("redis.addr")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return client
}
