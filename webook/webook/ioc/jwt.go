package ioc

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web/jwt"
	"github.com/redis/go-redis/v9"
	"time"
)

func InitRedisJWTHander(cmd redis.Cmdable) jwt.JwtHandler {
	atKey := []byte("Nj==FC5ZMTJncg@&fg!#%7aL#XNmemBZ")
	rtKey := []byte("quM@wyQR$DkHa6TBJ8acLSDYh4c2!@K5")
	accessExpireTime := time.Hour * 24
	refreshExpireTime := time.Hour * 24 * 7

	return jwt.NewRedisJwtHandler(cmd, atKey, rtKey, refreshExpireTime, accessExpireTime)
}
