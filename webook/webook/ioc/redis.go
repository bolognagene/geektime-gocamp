package ioc

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/key_expired_event"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/redisx"
	"github.com/prometheus/client_golang/prometheus"
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

	opts := prometheus.SummaryOpts{
		Namespace: "Golang",
		Subsystem: "Webook",
		Name:      "redis_resp_time",
		Help:      "Redis 相应时间",
	}
	redisx.NewPrometheusHook(opts)

	return client
}

func NewKeyExpiredKeys(k1 *key_expired_event.TopLikeKey) []key_expired_event.KeyExpiredEvent {
	return []key_expired_event.KeyExpiredEvent{k1}
}
