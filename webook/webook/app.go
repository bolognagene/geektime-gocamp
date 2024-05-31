package main

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/events"
	schedule_service "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/redisx"
	"github.com/gin-gonic/gin"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
	rh        *redisx.Handler
	// 分布式锁来计算Ranking
	//cron *cron.Cron
	//rankJob   *job.RankingJob
	// Scheduler
	cronJobScheduler *schedule_service.CronJobScheduler
}
