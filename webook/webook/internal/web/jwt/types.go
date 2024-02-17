package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JwtHandler interface {
	SetLoginToken(ctx *gin.Context, uid int64) error
	SetJWTToken(ctx *gin.Context, uid int64, ssid string) error
	SetRefreshJWTToken(ctx *gin.Context, uid int64, ssid string) error
	ClearToken(ctx *gin.Context) error
	CheckSession(ctx *gin.Context, ssid string) bool
	ExtractToken(ctx *gin.Context) string
	GetUserClaim(ctx *gin.Context) *UserClaims
	GetAtKey(ctx *gin.Context) []byte
}

type RefreshClaims struct {
	Uid  int64
	Ssid string
	jwt.RegisteredClaims
}

type UserClaims struct {
	Uid       int64
	Ssid      string
	UserAgent string
	jwt.RegisteredClaims
}
