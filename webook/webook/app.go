package main

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/events"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/job"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/redisx"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
	rh        *redisx.Handler
	cron      *cron.Cron
	rankJob   *job.RankingJob
}
