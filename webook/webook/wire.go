//go:build wireinject

package main

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/ioc"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/web"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func initWebServer() *gin.Engine {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB,
		ioc.InitRedis,

		// 初始化 DAO
		dao.NewUserDAO,

		// 初始化 Cache、
		cache.NewUserCache,
		// 用Redis方式存储
		ioc.InitSMSCodeRedisCache,

		// 初始化 Repository
		repository.NewUserRepository,
		repository.NewSMSCodeRepository,

		// 初始化 Service
		service.NewUserService,
		service.NewSMSCodeService,
		// Memory实现方式
		ioc.InitMemorySMSService,

		// 初始化 Handler
		web.NewUserHandler,

		// 中间件、路由等需要自己写一个函数放在这里
		ioc.InitWebServer,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
