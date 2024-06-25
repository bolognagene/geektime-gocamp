//go:build wireinject

package main

import (
	repository2 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interacitve/repository"
	cache2 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interacitve/repository/cache"
	dao2 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interacitve/repository/dao"
	service2 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interacitve/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interactive/events"
	event_article "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/events/article"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/key_expired_event"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao/article"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/ioc"
	schedule_repo "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/repository"
	schedule_dao "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/repository/dao"
	schedule_service "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/redisx"
	"github.com/google/wire"
	"time"
)

var interactiveSvcProvider = wire.NewSet(
	service2.NewInteractiveService,
	repository2.NewCachedInteractiveRepository,
	dao2.NewGORMInteractiveDAO,
	cache2.NewRedisInteractiveCache,
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
	ioc.InitRedisLoadSortCache,
)

// 用于mysql任务调度的实现方式
var cronJobSchedulerSet = wire.NewSet(
	ioc.InitCronJobScheduler,
	ioc.InitLocalFuncExecutor,
)

var cronJobSvcProvider = wire.NewSet(
	wire.Value(time.Duration(time.Minute)),
	schedule_service.NewPreemptCronJobService,
	schedule_repo.NewPreemptCronJobRepository,
	schedule_dao.NewGORMCronJobDAO,
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
		events.NewInteractiveReadEventConsumer,

		// redis key expired notify
		wire.Value(string("article")),
		key_expired_event.NewTopLikeKey,
		ioc.NewKeyExpiredKeys,
		redisx.NewHandler,
		// 分布式锁
		//ioc.InitRLockClient,

		// Service
		interactiveSvcProvider,
		articleServiceSet,
		rankingServiceSet,
		codeSvcProvider,
		userServiceSet,
		// cronjob scheduler
		cronJobSvcProvider,
		cronJobSchedulerSet,

		ioc.InitWechatService,
		// 直接基于内存实现
		ioc.InitSMSService,
		// 用于分布式锁的实现方式
		//ioc.InitRankingJob,
		//ioc.InitJobs,

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
