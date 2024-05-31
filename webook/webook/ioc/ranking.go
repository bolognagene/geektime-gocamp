package ioc

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/job"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"time"
)

func InitRedisLoadSortCache(client redis.Cmdable) *cache.RedisLoadSortCache {

	return cache.NewRedisLoadSortCache(client, "cronjob:ranking:topN:instance:load",
		time.Minute*3)
}

func InitRankingJob(svc service.RankingService,
	client *rlock.Client, l logger.Logger, loadCache *cache.RedisLoadSortCache) *job.RankingJob {
	return job.NewRankingJob(svc, time.Second*30,
		client, "cronjob:ranking:topN", l, loadCache, "instance1")
}

func InitJobs(l logger.Logger, rankingJob *job.RankingJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	adapter := job.NewRankingJobAdapter(rankingJob, l)
	// 这里每三分钟一次
	//_, err := res.AddJob("0 */3 * * * ?", adapter)
	_, err := res.AddJob("0 * * * * ?", adapter)
	if err != nil {
		panic(err)
	}
	return res
}

func InitRedisRankingCache(client redis.Cmdable) *cache.RedisRankingCache {
	return cache.NewRedisRankingCache(client, "RankingTopN", time.Minute*10)
}

func InitLocalRankingCache() *cache.LocalRankingCache {
	return cache.NewLocalRankingCache(time.Minute * 10)
}
