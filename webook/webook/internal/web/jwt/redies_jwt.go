package jwt

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

/*var (
	AtKey             = []byte("Nj==FC5ZMTJncg@&fg!#%7aL#XNmemBZ")
	RtKey             = []byte("quM@wyQR$DkHa6TBJ8acLSDYh4c2!@K5")
	refreshExpireTime = time.Hour * 24 * 7
	accessExpireTime  = time.Minute * 30
)*/

type RedisJwtHandler struct {
	cmd               redis.Cmdable
	AtKey             string
	RtKey             string
	refreshExpireTime time.Duration
	accessExpireTime  time.Duration
}

func NewRedisJwtHandler(cmd redis.Cmdable, atKey string, rtKey string, refreshExpireTime time.Duration, accessExpireTime time.Duration) *RedisJwtHandler {
	return &RedisJwtHandler{
		cmd:               cmd,
		AtKey:             atKey,
		RtKey:             rtKey,
		refreshExpireTime: refreshExpireTime,
		accessExpireTime:  accessExpireTime,
	}
}

/*func NewRedisJwtHandler(cmd redis.Cmdable) JwtHandler {
	return &RedisJwtHandler{
		cmd: cmd,
	}
}*/

func (h *RedisJwtHandler) GetAtKey(ctx *gin.Context) string {
	return h.AtKey
}

func (h *RedisJwtHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New()
	err := h.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	err = h.SetRefreshJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return nil
}

func (h *RedisJwtHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := UserClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.accessExpireTime)),
		},
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(h.AtKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h *RedisJwtHandler) SetRefreshJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := RefreshClaims{
		Uid:  uid,
		Ssid: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.refreshExpireTime)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(h.RtKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (h *RedisJwtHandler) ClearToken(ctx *gin.Context) error {
	// 清除access_token和refresh_token，这样resp里的这两个token就为空，前端就拿不到token了
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")

	// 此时要设置redis里的ssid为空，到期时间和refresh_token的到期时间一致
	// claims是在登录过后加进去的
	claims := ctx.MustGet("claims").(*UserClaims)
	return h.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid), "", h.refreshExpireTime).Err()
}

func (h *RedisJwtHandler) CheckSession(ctx *gin.Context, ssid string) bool {
	cnt, _ := h.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	if cnt > 0 {
		return true
	}
	return false

}

func (h *RedisJwtHandler) ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")
	segs := strings.Split(tokenHeader, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}

func (h *RedisJwtHandler) GetUserClaim(ctx *gin.Context) *UserClaims {
	// JWT方式
	c, exist := ctx.Get("claims")
	if !exist {
		//ctx.AbortWithStatus(http.StatusUnauthorized)
		return nil
	}

	// ok 代表是不是 *UserClaims
	claims, ok := c.(*UserClaims)
	if !ok {
		// 你可以考虑监控住这里
		//ctx.String(http.StatusOK, "系统错误")
		return nil
	}
	//println(claims.Uid)
	return claims
}
