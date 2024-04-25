package ioc

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
	"github.com/redis/go-redis/v9"
	"time"
)

func InitUserCache(client redis.Cmdable) cache.UserCache {
	return cache.NewUserCache(client, time.Minute*15)
}
