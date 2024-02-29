// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

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
	"time"
)

import (
	_ "github.com/spf13/viper/remote"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	jwtHandler := ioc.InitRedisJWTHander(cmdable)
	logger := ioc.InitLogger()
	v := ioc.InitMiddlewares(cmdable, jwtHandler, logger)
	db := ioc.InitDB(logger)
	userDAO := dao.NewUserDAO(db)
	duration := _wireDurationValue
	userCache := cache.NewUserCache(cmdable, duration)
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
	articleService := service.NewArticleService(articleRepository)
	interactiveDAO := dao.NewGORMInteractiveDAO(db)
	interactiveCache := cache.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveRepository)
	articleHandler := web.NewArticleHandler(articleService, interactiveService, logger)
	engine := ioc.InitWebServer(v, userHandler, oAuth2WechatHandler, articleHandler)
	return engine
}

var (
	_wireDurationValue = time.Minute *
		15
)
