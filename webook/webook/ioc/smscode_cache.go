package ioc

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/cache/sms"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/cache/sms/local"
	smsRedis "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/cache/sms/redis"
	"github.com/coocood/freecache"
	"github.com/redis/go-redis/v9"
)

func InitSMSCodeRedisCache(client redis.Cmdable) sms.CodeCache {
	return smsRedis.NewCodeCache(client)
}

func InitSMSCodeLocalCache(cache freecache.Cache) sms.CodeCache {
	return local.NewCodeCache(cache)
}
