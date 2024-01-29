package middleware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// 用session的方式登录校验
type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) IgnorePath(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	// 用 Go 的方式编码解码
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		// 不需要登录校验的
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		if id == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 这里是续约session的操作，每次超过1分钟就自动续约（设置session的MaxAge）
		now := time.Now()
		updateTime := sess.Get("update_time")
		if updateTime == nil {
			// 登录后第一次刷新， 直接设置update_time
			sess.Set("update_time", now)
			err := sess.Save()
			if err != nil {
				panic(err)
			}
			return
		}

		// 有update_time，判断是否大于1分钟，大于就重新设置session的MaxAge
		updateTimeVal, ok := updateTime.(time.Time)
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		sess.Set("update_time", now)
		if err := sess.Save(); err != nil {
			panic(err)
		}
		if now.Sub(updateTimeVal) > time.Second*10 {
			sess.Options(sessions.Options{
				MaxAge: 60,
			})
		}
	}
}
