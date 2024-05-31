package cache

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	//go:embed lua/get_load.lua
	luaGetLoad string
	//go:embed lua/set_load.lua
	luaSetLoad string
	//go:embed lua/get_sorted_load.lua
	luaGetSortedLoad string
)

type LoadSortCache interface {
	SetLoad(ctx context.Context, instance string, load float64) error
	GetLoad(ctx context.Context, instance string) (float64, error)
	GetSortedLowLoad(ctx context.Context, lowN int64) ([]string, error)
	GetSortedHighLoad(ctx context.Context, highN int64) ([]string, error)
}

type RedisLoadSortCache struct {
	Client redis.Cmdable
	key    string // zset的key
	ttl    time.Duration
}

func NewRedisLoadSortCache(client redis.Cmdable,
	key string, ttl time.Duration) *RedisLoadSortCache {
	return &RedisLoadSortCache{
		Client: client,
		key:    key,
		ttl:    ttl,
	}
}

func (r *RedisLoadSortCache) SetLoad(ctx context.Context, instance string, load float64) error {
	now := time.Now()
	return r.Client.Eval(ctx, luaSetLoad,
		[]string{r.key},
		instance, load, now.Add(r.ttl).UnixMilli()).Err()
}

func (r *RedisLoadSortCache) GetLoad(ctx context.Context, instance string) (float64, error) {
	now := time.Now().UnixMilli()
	return r.Client.Eval(ctx, luaGetLoad,
		[]string{r.key},
		instance, now).Float64()
}

func (r *RedisLoadSortCache) GetSortedLowLoad(ctx context.Context, lowN int64) ([]string, error) {
	//return r.Client.ZRange(ctx, r.key, 0, lowN).Result()
	now := time.Now().UnixMilli()
	return r.Client.Eval(ctx, luaGetSortedLoad,
		[]string{r.key},
		lowN, now, 0).StringSlice()
}

func (r *RedisLoadSortCache) GetSortedHighLoad(ctx context.Context, highN int64) ([]string, error) {
	//return r.Client.ZRevRange(ctx, r.key, 0, highN).Result()
	now := time.Now().UnixMilli()
	return r.Client.Eval(ctx, luaGetSortedLoad,
		[]string{r.key},
		highN, now, 1).StringSlice()
}
