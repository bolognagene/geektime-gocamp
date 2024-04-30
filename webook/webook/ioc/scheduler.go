package ioc

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
)

func InitLocalFuncExecutor(l logger.Logger) service.Executor {
	return service.NewLocalFuncExecutor(l)
}

func InitCronJobScheduler(svc service.CronJobService, l logger.Logger) service.JobScheduler {
	return service.NewCronJobScheduler(svc, l)
}
