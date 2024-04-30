package service

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"time"
)

type JobScheduler interface {
	Schedule(ctx context.Context) error
}

// CronJobScheduler Scheduler 调度器
type CronJobScheduler struct {
	execs map[string]Executor
	svc   CronJobService
	l     logger.Logger
}

func NewCronJobScheduler(svc CronJobService, l logger.Logger) JobScheduler {
	return &CronJobScheduler{
		svc:   svc,
		l:     l,
		execs: make(map[string]Executor),
	}
}

func (s *CronJobScheduler) Schedule(ctx context.Context) error {
	for {
		// 一次调度的数据库查询时间
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		job, err := s.svc.Preempt(dbCtx)
		cancel()

		if err != nil {
			// 你不能 return
			// 你要继续下一轮
			s.l.Debug("抢占任务失败", logger.Error(err))
		}

		exec, ok := s.execs[job.ExecutorName]
		if !ok {
			// DEBUG 的时候最好中断
			// 线上就继续
			s.l.Error("未找到对应的执行器",
				logger.String("executor_name", job.ExecutorName))
			continue
		}

		// 开始执行
		go func() {
			// 异步执行，不要阻塞主调度循环
			// 执行完毕之后
			// 这边要考虑超时控制，任务的超时控制
			err1 := exec.Exec(ctx, job)
			if err1 != nil {
				// 你也可以考虑在这里重试
				s.l.Error("任务执行失败", logger.Error(err1))
			}

			// 你要不要考虑下一次调度？设置next_time
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err1 = s.svc.ResetNextTime(ctx, job)
		}()
	}
}
