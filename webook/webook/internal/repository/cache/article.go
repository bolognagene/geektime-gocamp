package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type ArticleCache interface {
	Get(ctx context.Context, id int64) (domain.Article, error)
	Set(ctx context.Context, article domain.Article, time time.Duration) error
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, data []domain.Article) error
	DelFirstPage(ctx context.Context, uid int64) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	SetPub(ctx context.Context, article domain.Article, time time.Duration) error
}

type RedisArticleCache struct {
	client redis.Cmdable
}

func NewRedisArticleCache(client redis.Cmdable) ArticleCache {
	return &RedisArticleCache{
		client: client,
	}
}

func (r *RedisArticleCache) Get(ctx context.Context, id int64) (domain.Article, error) {
	// 可以直接使用 Bytes 方法来获得 []byte
	data, err := r.client.Get(ctx, r.articleIdKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(data, &res)
	return res, err

}

func (r *RedisArticleCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	// 可以直接使用 Bytes 方法来获得 []byte
	data, err := r.client.Get(ctx, r.publishArticleIdKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(data, &res)
	return res, err

}

func (r *RedisArticleCache) Set(ctx context.Context, article domain.Article, time time.Duration) error {
	data, err := json.Marshal(article)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.articleIdKey(article.Id), data, time).Err()
}

func (r *RedisArticleCache) SetPub(ctx context.Context, article domain.Article, time time.Duration) error {
	data, err := json.Marshal(article)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.publishArticleIdKey(article.Id), data, time).Err()
}

func (r *RedisArticleCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	data, err := r.client.Get(ctx, r.firstPageKey(uid)).Bytes()
	if err != nil {
		return nil, err
	}
	var articles []domain.Article
	err = json.Unmarshal(data, &articles)
	return articles, err
}

func (r *RedisArticleCache) SetFirstPage(ctx context.Context, uid int64, arts []domain.Article) error {
	for i := range arts {
		// 只缓存摘要部分
		arts[i].Content = arts[i].Abstract()
	}

	data, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	err = r.client.Set(ctx, r.firstPageKey(uid), data, time.Minute*30).Err()

	return err
}

func (r *RedisArticleCache) DelFirstPage(ctx context.Context, uid int64) error {
	return r.client.Del(ctx, r.firstPageKey(uid)).Err()
}

func (r *RedisArticleCache) firstPageKey(uid int64) string {
	return fmt.Sprintf("firstpage:%d", uid)
}

func (r *RedisArticleCache) articleIdKey(id int64) string {
	return fmt.Sprintf("article:%d", id)
}

func (r *RedisArticleCache) publishArticleIdKey(id int64) string {
	return fmt.Sprintf("published_article:%d", id)
}
