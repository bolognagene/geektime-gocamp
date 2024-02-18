package ginx

import (
	myjwt "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web/jwt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func WrapToken[C any](fn func(ctx *gin.Context, uc C) (Result, error), method string, l logger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var valAssert C
		var ok bool
		var val any
		switch any(valAssert).(type) {
		case myjwt.UserClaims:
			val, ok = ctx.Get("claims")

		case myjwt.RefreshClaims:
			val, ok = ctx.Get("refresh-claims")

		default:
			l.Debug("未授权", logger.String("method", method))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if !ok {
			l.Debug("未授权", logger.String("method", method))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok := val.(*C)
		if !ok {
			l.Debug("未授权", logger.String("method", method))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 下半段的业务逻辑从哪里来？
		// 我的业务逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx, *c)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			l.Debug("处理业务逻辑出错", logger.String("method", method), logger.Error(err))
		} else {
			l.Info("处理业务逻辑成功", logger.String("method", method))
		}

		if res.Msg == strconv.Itoa(http.StatusUnauthorized) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}

		ctx.JSON(http.StatusOK, res)
		// 再执行一些东西
	}
}

func WrapBodyAndToken[Req any, C any](fn func(ctx *gin.Context, req Req, uc C) (Result, error), method string, l logger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			l.Debug("绑定Req失败", logger.String("method", method), logger.Error(err), logger.Any("Req", req))
			return
		}

		var valAssert C
		var ok bool
		var val any
		switch any(valAssert).(type) {
		case myjwt.UserClaims:
			val, ok = ctx.Get("claims")

		case myjwt.RefreshClaims:
			val, ok = ctx.Get("refresh-claims")

		default:
			l.Debug("未授权", logger.String("method", method))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if !ok {
			l.Debug("未授权", logger.String("method", method))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok := any(val).(*C)
		if !ok {
			l.Debug("未授权", logger.String("method", method))
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 下半段的业务逻辑从哪里来？
		// 我的业务逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx, req, *c)
		if err != nil {
			// 开始处理 error，其实就是记录一下日志
			l.Debug("处理业务逻辑出错", logger.String("method", method), logger.Error(err))
		} else {
			l.Info("处理业务逻辑成功", logger.String("method", method))
		}

		if res.Msg == strconv.Itoa(http.StatusUnauthorized) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}

		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBody[T any](fn func(ctx *gin.Context, req T) (Result, error), method string, l logger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T

		err := ctx.Bind(&req)
		if err != nil {
			l.Debug("绑定Req失败", logger.String("method", method), logger.Error(err), logger.Any("Req", req))
			return
		}

		// 下半段的业务逻辑从哪里来？
		// 我的业务逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx, req)
		if err != nil {
			l.Debug("处理业务逻辑出错", logger.String("method", method), logger.Error(err))
		} else {
			l.Info("处理业务逻辑成功", logger.String("method", method))
		}

		if res.Msg == strconv.Itoa(http.StatusUnauthorized) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}

		ctx.JSON(http.StatusOK, res)
	}
}

func WrapFunc(fn func(ctx *gin.Context) (Result, error), method string, l logger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 下半段的业务逻辑从哪里来？
		// 我的业务逻辑有可能要操作 ctx
		// 你要读取 HTTP HEADER
		res, err := fn(ctx)
		if err != nil {
			l.Debug("处理业务逻辑出错", logger.String("method", method), logger.Error(err))
		} else {
			l.Info("处理业务逻辑成功", logger.String("method", method))
		}

		if res.Msg == strconv.Itoa(http.StatusUnauthorized) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}

		ctx.JSON(http.StatusOK, res)
	}
}

type Result struct {
	// 这个叫做业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
