package main

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/config"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/dao/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/web"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	db := initDB()
	redisClient := initRedis()
	server := initWebServer()

	u := initUser(db, redisClient)
	u.RegisterRoutes(server)

	/*server := gin.Default()
		server.GET("/hello", func(ctx *gin.Context) {
			ctx.String(http.StatusOK, "Hello!")
	})*/

	server.Run(":8081")
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	/*server.Use(func(ctx *gin.Context) {
		println("这是第一个 middleware")
	})

	server.Use(func(ctx *gin.Context) {
		println("这是第二个 middleware")
	})*/

	server.Use(cors.New(cors.Config{
		//AllowOrigins: []string{"*"},
		//AllowMethods: []string{"POST", "GET"},
		AllowHeaders:  []string{"Content-Type", "Authorization"},
		ExposeHeaders: []string{"x-jwt-token"},
		// 是否允许你带 cookie 之类的东西
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 你的开发环境
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	// 步骤1
	//store := cookie.NewStore([]byte("secret"))
	//server.Use(sessions.Sessions("webook", store))
	// 步骤3
	/*server.Use(middleware.NewLoginMiddlewareBuilder().
	IgnorePaths("/users/signup").
	IgnorePaths("/users/login").Build())*/

	// v1
	//middleware.IgnorePaths = []string{"sss"}
	//server.Use(middleware.CheckLogin())

	// 不能忽略sss这条路径
	//server1 := gin.Default()
	//server1.Use(middleware.CheckLogin())

	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").Build())
	return server
}

func initUser(db *gorm.DB, redisClient *redis.Client) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	uc := cache.NewUserCache(redisClient)
	repo := repository.NewUserRepository(ud, uc)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		// 我只会在初始化过程中 panic
		// panic 相当于整个 goroutine 结束
		// 一旦初始化过程出错，应用就不要启动了
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initRedis() *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})

	return redisClient
}
