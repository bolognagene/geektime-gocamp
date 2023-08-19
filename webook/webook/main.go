package main

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/service"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/web"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/web/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
)

func main() {
	db := initDB()
	server := initWebServer()

	u := initUser(db)
	u.RegisterRoutes(server)

	up := initUserProfile(db)
	up.RegisterRoutes(server)

	/*server := gin.Default()
		server.GET("/hello", func(ctx *gin.Context) {
			ctx.String(http.StatusOK, "Hello!")
	})*/

	server.Run(":8077")
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
		AllowHeaders: []string{"Content-Type", "Authorization"},
		//ExposeHeaders: []string{"x-jwt-token"},
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
	store := cookie.NewStore([]byte("secret"))
	server.Use(sessions.Sessions("webook", store))
	// 步骤3
	server.Use(middleware.NewLoginMiddlewareBuilder().
		IgnorePaths("/users/signup").
		IgnorePaths("/users/login").Build())

	// v1
	//middleware.IgnorePaths = []string{"sss"}
	//server.Use(middleware.CheckLogin())

	// 不能忽略sss这条路径
	//server1 := gin.Default()
	//server1.Use(middleware.CheckLogin())
	return server
}

func initUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}

func initUserProfile(db *gorm.DB) *web.UserProfileHandler {
	upd := dao.NewUserProfileDAO(db)
	repo := repository.NewUserProfileRepository(upd)
	svc := service.NewUserProfileService(repo)
	up := web.NewUserProfileHandler(svc)
	return up
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:12345678@tcp(192.168.181.130:3306)/webook"))
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
