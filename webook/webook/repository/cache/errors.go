package cache

import (
	"errors"
	"github.com/redis/go-redis/v9"
)

var (
	ErrKeyNotExist            = redis.Nil
	ErrCodeSendTooMany        = errors.New("发送验证码太频繁")
	ErrCodeVerifyTooManyTimes = errors.New("验证次数太多")
	ErrUnknownForCode         = errors.New("我也不知发生什么了，反正是跟 code 有关")
	ErrSystemError            = errors.New("系统错误")
)
