package repository

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/dao"
	"github.com/redis/go-redis/v9"
)

var (
	ErrUserDuplicateEmail     = dao.ErrUserDuplicateEmail
	ErrUserNotFound           = dao.ErrUserNotFound
	ErrKeyNotExist            = redis.Nil
	ErrCodeSendTooMany        = cache.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
	ErrUnknownForCode         = cache.ErrUnknownForCode
	ErrSystemError            = cache.ErrSystemError
)
