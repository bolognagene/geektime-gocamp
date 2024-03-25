//go:build wireinject

package main

import (
	event_article "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/events/article"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao/article"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/ioc"
	"github.com/google/wire"
	"time"
)

func InitWebServer() *App {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,
		// For MongoDB
		//ioc.InitMongoDB,
		//ioc.InitSnowflakeNode,
		// For kafka
		ioc.InitKafka,
		ioc.NewSyncProducer,
		ioc.NewConsumers,

		// consumer & producer
		event_article.NewKafkaProducer,
		event_article.NewInteractiveReadEventBatchConsumer,

		// 初始化 DAO
		dao.NewUserDAO,
		article.NewGORMArticleDAO,
		//article.NewMongoArticle,
		dao.NewGORMInteractiveDAO,

		wire.Value(time.Minute*15),
		cache.NewUserCache,
		cache.NewCodeCache,
		cache.NewRedisArticleCache,
		cache.NewRedisInteractiveCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedArticleRepository,
		repository.NewCachedInteractiveRepository,

		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		ioc.InitWechatService,
		// 直接基于内存实现
		ioc.InitSMSService,
		service.NewInteractiveService,

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
		// 组装我这个结构体的所有字段
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
