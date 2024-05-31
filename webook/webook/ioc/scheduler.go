package ioc

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/domain"
	schedulerSvc "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"time"
)

/*func InitLocalFuncExecutor(l logger.Logger) service.Executor {
	return service.NewLocalFuncExecutor(l)
}*/

/*func InitCronJobScheduler(svc service.CronJobService, l logger.Logger) service.JobScheduler {
	res := service.NewCronJobScheduler(svc, l)
	res.
}*/

func InitCronJobScheduler(l logger.Logger,
	local *schedulerSvc.LocalFuncExecutor,
	svc schedulerSvc.CronJobService) *schedulerSvc.CronJobScheduler {
	res := schedulerSvc.NewCronJobScheduler(svc, l)
	res.RegisterExecutor(local)
	return res
}

func InitLocalFuncExecutor(svc service.RankingService,
	l logger.Logger) *schedulerSvc.LocalFuncExecutor {
	res := schedulerSvc.NewLocalFuncExecutor(l)
	// 要在数据库里面插入一条记录。 手动插入RankingJob的记录
	// ranking job 的记录，通过管理任务接口来插入
	res.RegisterFunc("ranking", func(ctx context.Context, j domain.Job) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		return svc.TopN(ctx)
	})

	return res
}
