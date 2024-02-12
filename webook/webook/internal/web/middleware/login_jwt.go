package middleware

import (
	myjwt "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

// 用JWT的方式登录校验
type LoginJWTMiddlewareBuilder struct {
	paths []string
	myjwt.JwtHandler
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePath(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要登录校验的
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		// 用JWT的方式来校验
		// 得到请求头里的Authorization， 一般是 Bearer *****的格式
		tokenStr := l.ExtractToken(ctx)
		// 这里定义claims为指针，因为ParseWithClaims函数需要修改这个claims
		claims := &myjwt.UserClaims{}
		/*token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("TXqESPLch4roEwRPzo0WOkvGhpW4y0FU"), nil
		})*/
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return l.GetAtKey(ctx), nil
		})
		if err != nil {
			// 没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims.UserAgent != ctx.Request.UserAgent() {
			// UserAgent不一致，有安全隐患，需要重新登录
			// 加监控
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// err 为 nil，token 不为 nil
		if token == nil || !token.Valid || claims.Uid == 0 {
			// 没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 这里要在Redis里查询ssid这个key是否存在, 如果有证明已经登出
		if l.CheckSession(ctx, claims.Ssid) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 你以为的退出登录，没有用的
		//token.Valid = false
		//// tokenStr 是一个新的字符串
		//tokenStr, err = token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
		//if err != nil {
		//	// 记录日志
		//	log.Println("jwt 续约失败", err)
		//}
		//ctx.Header("x-jwt-token", tokenStr)

		// 短的 token 过期了，搞个新的
		//now := time.Now()
		// 每十秒钟刷新一次
		//if claims.ExpiresAt.Sub(now) < time.Second*50 {
		//	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
		//	tokenStr, err = token.SignedString([]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"))
		//	if err != nil {
		//		// 记录日志
		//		log.Println("jwt 续约失败", err)
		//	}
		//	ctx.Header("x-jwt-token", tokenStr)
		//}

		// JWT 方式需要将claims存到context里，供后续调用
		ctx.Set("claims", claims)
	}
}
