package redis

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/cache"
	"github.com/redis/go-redis/v9"
)

// 编译器会在编译的时候 把 set_code 的代码放进来这个 luaSetCode 变量里
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) *CodeCache {
	//return &sms.CodeCache{client: client}
	return &CodeCache{
		client: client,
	}
}

func (c *CodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.Key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}

	switch res {
	case 0:
		// 毫无问题
		return nil
	case -1:
		// 发送太频繁
		return cache.ErrCodeSendTooMany
	//case -2:
	//	return
	default:
		// 系统错误
		return cache.ErrSystemError
	}
}

func (c *CodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	res, err := c.client.Eval(ctx, luaVerifyCode, []string{c.Key(biz, phone)}, code).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		// 正常来说，如果频繁出现这个错误，你就要告警，因为有人搞你
		return false, cache.ErrCodeVerifyTooManyTimes
	case -2:
		return false, nil
		//default:
		//	return false, ErrUnknownForCode
	}
	return false, cache.ErrUnknownForCode

}

func (c *CodeCache) Key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
