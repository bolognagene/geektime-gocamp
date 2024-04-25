package ioc

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/config"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web/jwt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web/middleware"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/ginx"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/ginx/middlewares/logger"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/ginx/middlewares/metrics"
	logger2 "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	redissession "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler,
	oauth2wechatHdl *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	oauth2wechatHdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable, jwtHandler jwt.JwtHandler, l logger2.Logger) []gin.HandlerFunc {
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "Golang",
		Subsystem: "Webook",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	return []gin.HandlerFunc{
		corsHdl(),
		logger.NewBuilder(func(ctx context.Context, al *logger.AccessLog) {
			l.Debug("HTTP请求", logger2.Field{
				Key:   "AccessLog",
				Value: al,
			})
		}).AllowReqBody().AllowRespBody().AllowEncrypt().CountLimit(1024).Build(),
		metrics.NewBuilder("Golang",
			"Webook", "gin_http", "统计GIN的HTTP接口", "myinstance").Build(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHandler).
			IgnorePath("/users/signup").
			IgnorePath("/users/login_sms/code/send").
			IgnorePath("/users/login_sms").
			IgnorePath("/users/login").Build(),
		//ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
		setJWTToken(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		//AllowOrigins: []string{"*"},
		//AllowMethods: []string{"POST", "GET"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 你不加这个，前端是拿不到的
		ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
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
	})

}

func setJWTToken() gin.HandlerFunc {
	store, err := redissession.NewStore(16, "tcp", config.Config.Redis.Addr,
		"", []byte("B8tORKiMxPU4MkhkFTXsG0BSQO5D0WwW"), []byte("FAmJ7sk504aTKkZrhDFpz4PekpeO3jOg"))
	if err != nil {
		panic(err)
	}

	return sessions.Sessions("genesession", store)
}
