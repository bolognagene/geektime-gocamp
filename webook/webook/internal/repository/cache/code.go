package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (
	ErrCodeSendTooMany        = errors.New("发送验证码太频繁")
	ErrCodeVerifyTooManyTimes = errors.New("验证次数太多")
	ErrUnknownForCode         = errors.New("我也不知发生什么了，反正是跟 code 有关")
)

// 编译器会在编译的时候，把 set_code 的代码放进来这个 luaSetCode 变量里
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) *CodeCache {
	return &CodeCache{
		client: client,
	}
}

func (cache *CodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := cache.client.Eval(ctx, luaSetCode, []string{cache.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		// 毫无问题
		return nil
	case -1:
		// 发送太频繁
		return ErrCodeSendTooMany
	default:
		// 系统错误
		return errors.New("系统错误")

	}

}

func (cache *CodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	res, err := cache.client.Eval(ctx, luaVerifyCode, []string{cache.key(biz, phone)}, code).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		// 毫无问题
		return true, nil
	case -1:
		// 发送太频繁
		return false, ErrCodeVerifyTooManyTimes
	case -2:
		return false, nil
	}
	return false, ErrUnknownForCode
}

func (cache *CodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
