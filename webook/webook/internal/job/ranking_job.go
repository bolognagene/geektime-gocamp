package job

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	rlock "github.com/gotomicro/redis-lock"
	"math/rand"
	"sync"
	"time"
)

type RankingJob struct {
	svc       service.RankingService
	timeout   time.Duration
	client    *rlock.Client
	lock      *rlock.Lock
	key       string
	localLock *sync.Mutex
	l         logger.Logger
	loadCache cache.RedisLoadSortCache // 拿锁时增加对Load的检测
	instance  string                   // 标记是哪个instance
}

func NewRankingJob(svc service.RankingService,
	timeout time.Duration,
	client *rlock.Client,
	key string,
	l logger.Logger,
	loadCache cache.RedisLoadSortCache,
	instance string) *RankingJob {
	return &RankingJob{svc: svc,
		timeout:   timeout,
		client:    client,
		key:       key,
		l:         l,
		localLock: &sync.Mutex{},
		loadCache: loadCache,
		instance:  instance,
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

// Run 按时间调度的，三分钟一次
func (r *RankingJob) Run() error {
	// 这个localLock是用来保证if r.lock == nil 这句话的
	// 比如一个进程刚设置好r.lock = lock, 结果与此同时AutoRefresh的goroutine又给设置成r.lock = nil
	r.localLock.Lock()
	defer r.localLock.Unlock()
	// 这里将此时的Load数据写入Cache
	// 随机数模拟Load
	load := getLoad()
	err := r.loadCache.SetLoad(context.Background(), r.instance, load)
	if err != nil {
		return err
	}

	lowLoadInstances, err := r.loadCache.
		GetSortedLowLoad(context.Background(), 2)
	// 如果有err，那还是参与拿锁
	if err == nil && lowLoadInstances != nil {
		// 当前instance的负载不在最低的3个，
		// 如果当前有锁，则还要Unlock
		if !slice.Contains[string](lowLoadInstances, r.instance) {
			r.l.Debug("当前instance负载不为最小的前三个",
				logger.String("instance", r.instance),
				logger.Float64("load", load))
			if r.lock != nil {
				r.lock.Unlock(context.Background())
			}
			return nil
		}
	}

	if r.lock == nil {
		// 说明你没拿到锁，你得试着拿锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 我可以设置一个比较短的过期时间，这是设置为30秒
		// 比如你的定时任务是3分钟运行一次，那么这里你可以设置r.timeout为一个 比较短的时间， 比如1分钟
		rlock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      0, // 这里设置重试次数为0，因为这里是试着拿锁，很有可能这个锁已经被其他instance拿到了
		}, time.Second)
		if err != nil {
			// 这边没拿到锁，极大概率是别人持有了锁
			return nil
		}

		r.lock = rlock
		// 我怎么保证我这里，一直拿着这个锁？？？
		// 续约
		go func() {
			// 这里有必要加锁吗？
			//r.localLock.Lock()
			//defer r.localLock.Unlock()
			// 自动续约机制
			// AutoRefresh里是一直以r.timeout/2的间隔来续约的
			err1 := rlock.AutoRefresh(r.timeout/2, time.Second)
			// 这里说明退出了续约机制
			// 续约失败了怎么办？
			if err1 != nil {
				// 不怎么办
				// 争取下一次，继续抢锁
				r.l.Error("续约失败", logger.Error(err))

			}
			// 为什么无论AutoRefresh返回什么都要设置r.lock=nil ？
			// 因为AutoRefresh里是一个无限循环，除非续约失败，否则会一直循环而不退出AutoRefresh函数
			// 所以一旦走到这里，表示AutoRefresh退出了，一定是续约失败了
			// 所以要把r.lock设置为nil
			r.lock = nil
			// lock.Unlock(ctx)
		}()
	}
	//拿到分布式锁了，可以执行了
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)

}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.lock = nil
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

// 模拟取得Load数， 随机生成
func getLoad() float64 {
	return rand.Float64() * 100
}
