package main

func main() {
	server := InitWebServer()

	server.Run(":8077") // 监听并在 0.0.0.0:8077 上启动服务
}

/*func initUser(db *gorm.DB, cache redis_client.Cmdable) *web.UserHandler {
	userDao := dao.NewUserDAO(db)
	userCache := mycache.NewUserCache(cache, time.Minute*15)
	userRepo := repository.NewUserRepository(userDao, userCache)
	codeCache := mycache.NewCodeCache(cache)
	codeRepo := repository.NewCodeRepository(codeCache)
	memorySMSService := memory.NewService()
	codeService := service.NewCodeService(codeRepo, memorySMSService)
	userService := service.NewUserService(userRepo)
	userHandler := web.NewUserHandler(userService, codeService)
	return userHandler
}*/

/*func InitWebServer() *gin.Engine {
	server := gin.Default()

	return server
}*/
