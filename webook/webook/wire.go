//go:build wireinject

package main

import (
	event_article "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/events/article"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/key_expired_event"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao/article"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/ioc"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/redisx"
	"github.com/google/wire"
)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

var articleServiceSet = wire.NewSet(
	service.NewArticleService,
	repository.NewCachedArticleRepository,
	article.NewGORMArticleDAO,
	cache.NewRedisArticleCache,
	//article.NewMongoArticle, //MongoDB
	//article.NewOssDAO, //OSS
)

var rankingServiceSet = wire.NewSet(
	service.NewBatchRankingService,
	repository.NewCachedRankingRepository,
	ioc.InitLocalRankingCache,
	ioc.InitRedisRankingCache,
)

var userServiceSet = wire.NewSet(
	service.NewUserService,
	repository.NewUserRepository,
	dao.NewUserDAO,
	ioc.InitUserCache, //包含一个具体的时间，所以需要另写一个函数
)

var codeSvcProvider = wire.NewSet(
	service.NewCodeService,
	repository.NewCodeRepository,
	cache.NewCodeCache,
)

/*var cronJobSvcProvider = wire.NewSet(
	wire.Value(time.Duration(time.Minute)),
	schedule_service.NewPreemptCronJobService,
	schedule_repo.NewPreemptCronJobRepository,
	schedule_dao.NewGORMCronJobDAO,
)*/

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
		//event_article.NewInteractiveReadEventBatchConsumer,
		event_article.NewInteractiveReadEventConsumer,

		// redis key expired notify
		wire.Value(string("article")),
		key_expired_event.NewTopLikeKey,
		ioc.NewKeyExpiredKeys,
		redisx.NewHandler,
		ioc.InitRLockClient,

		// Service
		interactiveSvcProvider,
		articleServiceSet,
		rankingServiceSet,
		codeSvcProvider,
		userServiceSet,
		//cronJobSvcProvider,

		ioc.InitWechatService,
		// 直接基于内存实现
		ioc.InitSMSService,
		ioc.InitRankingJob,
		ioc.InitJobs,

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
