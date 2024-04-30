package job

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
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
}

func NewRankingJob(svc service.RankingService, timeout time.Duration, client *rlock.Client, key string, l logger.Logger) *RankingJob {
	return &RankingJob{svc: svc,
		timeout:   timeout,
		client:    client,
		key:       key,
		l:         l,
		localLock: &sync.Mutex{},
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	defer r.localLock.Unlock()
	if r.lock == nil {
		// 说明你没拿到锁，你得试着拿锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 我可以设置一个比较短的过期时间
		// 比如你的定时任务是3分钟运行一次，那么这里你可以设置r.timeout为一个 比较短的时间， 比如1分钟
		rlock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      0,
		}, time.Second)
		if err != nil {
			// 这边没拿到锁，极大概率是别人持有了锁
			return nil
		}

		r.lock = rlock
		// 我怎么保证我这里，一直拿着这个锁？？？
		// 续约
		go func() {
			r.localLock.Lock()
			defer r.localLock.Unlock()
			// 自动续约机制
			err1 := rlock.AutoRefresh(r.timeout/2, time.Second)
			// 这里说明退出了续约机制
			// 续约失败了怎么办？
			if err1 != nil {
				// 不怎么办
				// 争取下一次，继续抢锁
				r.l.Error("续约失败", logger.Error(err))

			}
			// 为什么无论AutoRefresh返回什么都要设置r.lock=nil ？
			// 1, 什么时候AutoRefresh函数会返回 nil ?
			// 两种情况：续约成功 或者 锁已经被unlock了
			// 所以这里就要考虑为什么续约成功还要将r.lock 设置为 nil (2)
			// 以及如果是锁unlock了返回nil，而这里又不设置r.lock为nil会怎样 (3)
			// 2, 什么续约成功还要将r.lock 设置为 nil
			// 这里设置为nil， 那么当下一次Job运行的时候，会再进入这个分支，会继续调用Lock函数和AutoRefresh函数
			// 此时Redis里如果没有锁 （锁已经过期）那么就是正常的抢占锁的流程
			// 如果还有锁，那么此种情况下，Lock会刷新一下过期时间并返回一个锁，再往下走开启goroutine，进行AutoRefresh
			r.lock = nil
			// lock.Unlock(ctx)
		}()
	}
	//拿到分布式锁了，可以执行了
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)

}
