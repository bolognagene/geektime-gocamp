package cache

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/pkg/errors"
	"time"
)

type LocalRankingCache struct {
	// 我用我的泛型封装
	// 你可以考虑直接使用 uber 的，或者 SDK 自带的
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time] // 下一次调度时间
	expiration time.Duration
}

func NewLocalRankingCache(expiration time.Duration) *LocalRankingCache {
	return &LocalRankingCache{
		topN:       atomicx.NewValue[[]domain.Article](),
		ddl:        atomicx.NewValueOf(time.Now()),
		expiration: expiration,
	}
}

func (cache *LocalRankingCache) Set(ctx context.Context, arts []domain.Article) error {
	// 也可以按照 id => Article 缓存
	cache.topN.Store(arts)
	ddl := time.Now().Add(cache.expiration)
	cache.ddl.Store(ddl)
	return nil
}

func (cache *LocalRankingCache) Get(ctx context.Context) ([]domain.Article, error) {
	arts := cache.topN.Load()
	if len(arts) == 0 || cache.ddl.Load().Before(time.Now()) {
		return nil, errors.New("本地缓存未命中")
	}
	return arts, nil
}

func (cache *LocalRankingCache) ForceGet(ctx context.Context) ([]domain.Article, error) {
	arts := cache.topN.Load()
	return arts, nil
}
