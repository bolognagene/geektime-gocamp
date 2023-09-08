// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/ioc"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/web"
	"github.com/gin-gonic/gin"
)

// Injectors from wire.go:

func initWebServer() *gin.Engine {
	cmdable := ioc.InitRedis()
	v := ioc.InitMiddlewares(cmdable)
	db := ioc.InitDB()
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := ioc.InitSMSCodeRedisCache(cmdable)
	smsCodeRepository := repository.NewSMSCodeRepository(codeCache)
	smsService := ioc.InitMemorySMSService()
	smsCodeService := service.NewSMSCodeService(smsCodeRepository, smsService)
	userHandler := web.NewUserHandler(userService, smsCodeService)
	engine := ioc.InitWebServer(v, userHandler)
	return engine
}