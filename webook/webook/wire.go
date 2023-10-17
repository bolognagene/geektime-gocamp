//go:build wireinject

package main

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web"
	ijwt "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web/jwt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,

		// 初始化 DAO
		dao.NewUserDAO,
		dao.NewSMSDAO,

		cache.NewUserCache,
		cache.NewCodeCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewSMSAsyncRepository,

		service.NewUserService,
		service.NewCodeService,
		// 直接基于内存实现
		ioc.InitSMSService,
		ioc.InitWechatService,

		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ioc.NewWechatHandlerConfig,
		ijwt.NewRedisJWTHandler,
		// 你中间件呢？
		// 你注册路由呢？
		// 你这个地方没有用到前面的任何东西
		//gin.Default,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
