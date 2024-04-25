package cache

import (
	"context"
	"encoding/json"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type RankingCache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}

type RedisRankingCache struct {
	client redis.Cmdable
	key    string
}

func NewRedisRankingCache(client redis.Cmdable, key string) *RedisRankingCache {
	return &RedisRankingCache{
		client: client,
		key:    key,
	}
}

func (r *RedisRankingCache) Set(ctx context.Context, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = "" //热榜无需保存文章内容，设为空节省redis空间
	}
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	// 这个过期时间要稍微长一点，最好是超过计算热榜的时间（包含重试在内的时间）
	// 你甚至可以直接永不过期
	return r.client.Set(ctx, r.key, val, time.Minute*10).Err()
}

func (r *RedisRankingCache) Get(ctx context.Context) ([]domain.Article, error) {
	data, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}

	var res []domain.Article
	err = json.Unmarshal(data, res)
	return res, err
}
