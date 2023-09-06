package local

import (
	"context"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/cache"
	"github.com/coocood/freecache"
	"strconv"
	"sync"
)

type CodeCache struct {
	cache freecache.Cache
}

func NewCodeCache(cache freecache.Cache) CodeCache {
	return CodeCache{
		cache: cache,
	}

}

func (c *CodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	var rwMutex sync.RWMutex
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	key := c.Key(biz, phone)
	keyCnt := key + "_cnt"
	keyExpire := 600

	cnt, err := c.cache.Get([]byte(keyCnt))
	if err != nil {
		return false, cache.ErrSystemError
	}

	count, err := strconv.Atoi(string(cnt))
	if err != nil {
		return false, cache.ErrSystemError
	}

	if count <= 0 {
		return false, cache.ErrCodeVerifyTooManyTimes
	}

	cacheCode, err := c.cache.Get([]byte(key))
	if err != nil && err != freecache.ErrNotFound {
		return false, cache.ErrSystemError
	}

	if err == freecache.ErrNotFound {
		return false, cache.ErrKeyNotExist
	}

	if string(cacheCode) == code {
		// 验证成功
		c.cache.Set([]byte(keyCnt), []byte("-1"), keyExpire)
		return true, nil
	} else {
		// 验证失败, count - 1
		count = count - 1
		c.cache.Set([]byte(keyCnt), []byte(strconv.Itoa(count)), keyExpire)
		return false, cache.ErrUnknownForCode
	}
}

func (c *CodeCache) Set(ctx context.Context, biz, phone, code string) error {
	var rwMutex sync.RWMutex
	rwMutex.RLock()
	defer rwMutex.RUnlock()

	key := c.Key(biz, phone)
	keyCnt := key + "_cnt"
	keyExpire := 600

	ttl, err := c.cache.TTL([]byte(key))
	if err != nil && err != freecache.ErrNotFound {
		return cache.ErrSystemError
	}

	if err == freecache.ErrNotFound || ttl < 540 {
		// 正常情况设置key
		c.cache.Set([]byte(key), []byte(code), keyExpire)
		c.cache.Set([]byte(keyCnt), []byte("3"), keyExpire)

		return nil
	} else {
		// 发送太频繁
		return cache.ErrCodeSendTooMany
	}

}

func (c *CodeCache) Key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
