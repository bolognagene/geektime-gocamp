package ginx

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web/jwt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"github.com/gin-gonic/gin"
)

func WrapReq[T any](fn func(ctx *gin.Context, req T, uc jwt.UserClaims) (Result, error), l logger.LoggerV1, method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 顺便把 userClaims 也取出来
		var req T
		var uc jwt.UserClaims
		if err := ctx.Bind(&req); err != nil {
			return
		}
		res, err := fn(ctx, req, uc)
		if err != nil {
			l.Error(method, logger.Field{
				Key:   "error",
				Value: err.Error(),
			}, logger.Field{
				Key:   "response",
				Value: res,
			})
		}
	}
}

type Result struct {
	// 这个叫做业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
