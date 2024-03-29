package main

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/events"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/redisx"
	"github.com/gin-gonic/gin"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
	rh        *redisx.Handler
}
