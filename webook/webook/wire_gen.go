// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interactive/events"
	repository3 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interactive/repository"
	cache2 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interactive/repository/cache"
	dao3 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interactive/repository/dao"
	service3 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interactive/service"
	article2 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/events/article"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/key_expired_event"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao/article"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/ioc"
	repository2 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/repository"
	dao2 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/repository/dao"
	service2 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/redisx"
	"github.com/google/wire"
	"time"
)

import (
	_ "github.com/spf13/viper/remote"
)

// Injectors from wire.go:

func InitWebServer() *App {
	cmdable := ioc.InitRedis()
	jwtHandler := ioc.InitRedisJWTHander(cmdable)
	logger := ioc.InitLogger()
	v := ioc.InitMiddlewares(cmdable, jwtHandler, logger)
	db := ioc.InitDB(logger)
	userDAO := dao.NewUserDAO(db)
	userCache := ioc.InitUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService(cmdable)
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, jwtHandler, logger)
	wechatService := ioc.InitWechatService()
	wechatHandlerConfig := ioc.NewWechatHandlerConfig()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, wechatHandlerConfig)
	articleDAO := article.NewGORMArticleDAO(db)
	articleCache := cache.NewRedisArticleCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(articleDAO, userDAO, articleCache, logger)
	client := ioc.InitKafka()
	syncProducer := ioc.NewSyncProducer(client)
	producer := article2.NewKafkaProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer, logger)
	interactiveDAO := dao3.NewGORMInteractiveDAO(db)
	interactiveCache := cache2.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository3.NewCachedInteractiveRepository(interactiveDAO, interactiveCache, cmdable, logger)
	interactiveService := service3.NewInteractiveService(interactiveRepository)
	articleHandler := web.NewArticleHandler(articleService, interactiveService, logger)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler)
	interactiveReadEventConsumer := events.NewInteractiveReadEventConsumer(client, interactiveRepository, logger)
	v2 := ioc.NewConsumers(interactiveReadEventConsumer)
	string2 := _wireStringValue
	topLikeKey := key_expired_event.NewTopLikeKey(interactiveRepository, logger, string2)
	v3 := ioc.NewKeyExpiredKeys(topLikeKey)
	handler := redisx.NewHandler(cmdable, v3)
	redisRankingCache := ioc.InitRedisRankingCache(cmdable)
	localRankingCache := ioc.InitLocalRankingCache()
	rankingRepository := repository.NewCachedRankingRepository(redisRankingCache, localRankingCache)
	rankingService := service.NewBatchRankingService(articleService, interactiveService, rankingRepository)
	localFuncExecutor := ioc.InitLocalFuncExecutor(rankingService, logger)
	cronJobDAO := dao2.NewGORMCronJobDAO(db)
	cronJobRepository := repository2.NewPreemptCronJobRepository(cronJobDAO)
	duration := _wireDurationValue
	cronJobService := service2.NewPreemptCronJobService(cronJobRepository, duration, logger)
	cronJobScheduler := ioc.InitCronJobScheduler(logger, localFuncExecutor, cronJobService)
	app := &App{
		web:              engine,
		consumers:        v2,
		rh:               handler,
		cronJobScheduler: cronJobScheduler,
	}
	return app
}

var (
	_wireStringValue   = string("article")
	_wireDurationValue = time.Duration(time.Minute)
)

// wire.go:

var interactiveSvcProvider = wire.NewSet(service3.NewInteractiveService, repository3.NewCachedInteractiveRepository, dao3.NewGORMInteractiveDAO, cache2.NewRedisInteractiveCache)

var articleServiceSet = wire.NewSet(service.NewArticleService, repository.NewCachedArticleRepository, article.NewGORMArticleDAO, cache.NewRedisArticleCache)

var rankingServiceSet = wire.NewSet(service.NewBatchRankingService, repository.NewCachedRankingRepository, ioc.InitLocalRankingCache, ioc.InitRedisRankingCache, ioc.InitRedisLoadSortCache)

// 用于mysql任务调度的实现方式
var cronJobSchedulerSet = wire.NewSet(ioc.InitCronJobScheduler, ioc.InitLocalFuncExecutor)

var cronJobSvcProvider = wire.NewSet(wire.Value(time.Duration(time.Minute)), service2.NewPreemptCronJobService, repository2.NewPreemptCronJobRepository, dao2.NewGORMCronJobDAO)

var userServiceSet = wire.NewSet(service.NewUserService, repository.NewUserRepository, dao.NewUserDAO, ioc.InitUserCache)

var codeSvcProvider = wire.NewSet(service.NewCodeService, repository.NewCodeRepository, cache.NewCodeCache)
