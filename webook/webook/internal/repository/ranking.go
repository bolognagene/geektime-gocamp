package repository

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type CachedRankingRepository struct {
	// 使用具体实现，可读性更好，对测试不友好，因为没有面向接口编程
	redisCache *cache.RedisRankingCache
	localCache *cache.LocalRankingCache
}

func NewCachedRankingRepository(redis *cache.RedisRankingCache, local *cache.LocalRankingCache) RankingRepository {
	return &CachedRankingRepository{
		redisCache: redis,
		localCache: local,
	}
}

func (c *CachedRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	// 先Set local， 在Set redis。 因为localcache基本不会出错
	c.localCache.Set(ctx, arts)
	err := c.redisCache.Set(ctx, arts)
	return err
}

func (c *CachedRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	arts, err := c.localCache.Get(ctx)
	if err == nil {
		return arts, nil
	}
	//读取local失败了，从redis里读取
	arts, err = c.redisCache.Get(ctx)
	if err == nil {
		return arts, nil
	}
	// 如果此时还是报错，则强制从local里读
	return c.localCache.ForceGet(ctx)
}
