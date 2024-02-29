//go:build wireinject

package main

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao/article"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"time"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,
		// For MongoDB
		//ioc.InitMongoDB,
		//ioc.InitSnowflakeNode,

		// 初始化 DAO
		dao.NewUserDAO,
		article.NewGORMArticleDAO,
		//article.NewMongoArticle,

		wire.Value(time.Minute*15),
		cache.NewUserCache,
		cache.NewCodeCache,
		cache.NewRedisArticleCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedArticleRepository,

		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		ioc.InitWechatService,
		// 直接基于内存实现
		ioc.InitSMSService,
		ioc.NewWechatHandlerConfig,
		ioc.InitRedisJWTHander,
		// handler
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		// 你中间件呢？
		// 你注册路由呢？
		// 你这个地方没有用到前面的任何东西
		//gin.Default,

		ioc.InitWebServer,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
