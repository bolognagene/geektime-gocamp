package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var (
	//go:embed lua/interative_incr_cnt.lua
	luaIncrCnt string
	//go:embed lua/increment_top_like.lua
	luaIncrTopLike string
)

const (
	fieldReadCnt    = "read_cnt"
	fieldCollectCnt = "collect_cnt"
	fieldLikeCnt    = "like_cnt"
)

// 方案1
// key1 => map[string]int

// 方案2
// key1_read_cnt => 10
// key1_collect_cnt => 11
// key1_like_cnt => 13
type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	BatchIncrReadCntIfPresent(ctx context.Context, biz string, bizIds []int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) (int, error)
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	GetCnt(ctx context.Context, biz string, id int64) (domain.Interactive, error)
	SetCnt(ctx context.Context, biz string, bizId int64, interactive domain.Interactive) error
	IncrTopLike(ctx context.Context, biz string, bizId int64, limit int64) (int, error)
	DecrTopLike(ctx context.Context, biz string, bizId int64, limit int64) error
	GetTopLike(ctx context.Context, biz string, n int64) ([]domain.TopWithScore, error)
	SetTopLike(ctx context.Context, biz string, intrs []domain.TopWithScore) error
}

type RedisInteractiveCache struct {
	client redis.Cmdable
}

func NewRedisInteractiveCache(client redis.Cmdable) InteractiveCache {

	return &RedisInteractiveCache{
		client: client,
	}
}

func (r *RedisInteractiveCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	// 拿到的结果，可能自增成功了，可能不需要自增（key不存在）
	// 你要不要返回一个 error 表达 key 不存在？
	//res, err := r.client.Eval(ctx, luaIncrCnt,
	//	[]string{r.key(biz, bizId)},
	//	// read_cnt +1
	//	"read_cnt", 1).Int()
	//if err != nil {
	//	return err
	//}
	//if res == 0 {
	// 这边一般是缓存过期了
	//	return errors.New("缓存中 key 不存在")
	//}
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		// read_cnt +1
		fieldReadCnt, 1).Err()
}

func (r *RedisInteractiveCache) BatchIncrReadCntIfPresent(ctx context.Context, biz string, bizIds []int64) error {
	keys := make([]string, 0, len(bizIds))
	for _, bizId := range bizIds {
		keys = append(keys, r.key(biz, bizId))
	}
	return r.client.Eval(ctx, luaIncrCnt,
		keys,
		fieldReadCnt, 1).Err()
}

func (r *RedisInteractiveCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) (int, error) {
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		fieldLikeCnt, 1).Int()
}

func (r *RedisInteractiveCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		fieldLikeCnt, -1).Err()
}

func (r *RedisInteractiveCache) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		fieldCollectCnt, 1).Err()
}

func (r *RedisInteractiveCache) DecrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		fieldCollectCnt, -1).Err()
}

func (r *RedisInteractiveCache) GetCnt(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	// 直接使用 HMGet，即便缓存中没有对应的 key，也不会返回 error
	//r.client.HMGet(ctx, r.key(biz, bizId),
	//	fieldCollectCnt, fieldLikeCnt, fieldReadCnt)
	// 所以你没有办法判定，缓存里面是有这个key，但是对应 cnt 都是0，还是说没有这个 key

	// 拿到 key 对应的值里面的所有的 field
	data, err := r.client.HGetAll(ctx, r.key(biz, bizId)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}

	if len(data) == 0 {
		// 缓存不存在，系统错误，比如说你的同事，手贱设置了缓存，但是忘记任何 fields
		return domain.Interactive{}, ErrKeyNotExist
	}

	// 理论上来说，这里不可能有 error
	collectCnt, _ := strconv.ParseInt(data[fieldCollectCnt], 10, 64)
	likeCnt, _ := strconv.ParseInt(data[fieldLikeCnt], 10, 64)
	readCnt, _ := strconv.ParseInt(data[fieldReadCnt], 10, 64)

	return domain.Interactive{
		CollectCnt: collectCnt,
		LikeCnt:    likeCnt,
		ReadCnt:    readCnt,
	}, err
}

func (r *RedisInteractiveCache) SetCnt(ctx context.Context, biz string, bizId int64, interactive domain.Interactive) error {
	key := r.key(biz, bizId)
	err := r.client.HMSet(ctx, key,
		fieldLikeCnt, interactive.LikeCnt,
		fieldCollectCnt, interactive.CollectCnt,
		fieldReadCnt, interactive.ReadCnt).Err()
	if err != nil {
		return err
	}
	return r.client.Expire(ctx, key, time.Minute*15).Err()
}

func (r *RedisInteractiveCache) IncrTopLike(ctx context.Context, biz string, bizId int64, limit int64) (int, error) {
	return r.client.Eval(ctx, luaIncrTopLike,
		[]string{fmt.Sprintf("top_like_%s", biz)},
		1, bizId, limit).Int()
}

func (r *RedisInteractiveCache) DecrTopLike(ctx context.Context, biz string, bizId int64, limit int64) error {
	return r.client.Eval(ctx, luaIncrTopLike,
		[]string{fmt.Sprintf("top_like_%s", biz)},
		-1, bizId, limit).Err()
}

func (r *RedisInteractiveCache) GetTopLike(ctx context.Context, biz string, n int64) ([]domain.TopWithScore, error) {
	data, err := r.client.ZRevRangeWithScores(ctx, fmt.Sprintf("top_like_%s", biz), 0, n).Result()
	if err == nil && len(data) != 0 {
		ts := make([]domain.TopWithScore, n)
		for i, z := range data {
			tws, _ := r.ToTopWithScore(z)
			ts[i] = tws
		}
		return ts, err
	}
	return nil, err
}

func (r *RedisInteractiveCache) SetTopLike(ctx context.Context, biz string, intrs []domain.TopWithScore) error {
	zs := make([]redis.Z, len(intrs))
	for i, intr := range intrs {
		zs[i] = r.ToRedisZ(intr)
	}

	zargs := redis.ZAddArgs{
		Members: zs,
	}

	err := r.client.ZAddArgs(ctx, fmt.Sprintf("top_like_%s", biz), zargs).Err()
	if err != nil {
		return err
	}

	return r.client.Expire(ctx, fmt.Sprintf("top_like_%s", biz), time.Minute*1).Err()
}

func (r *RedisInteractiveCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}

func (r *RedisInteractiveCache) ToTopWithScore(z redis.Z) (domain.TopWithScore, error) {
	member, err := strconv.ParseInt(z.Member.(string), 10, 64)
	if err != nil {
		return domain.TopWithScore{
			Score:  0,
			Member: 0,
		}, err
	}

	return domain.TopWithScore{
		Score:  z.Score,
		Member: member,
	}, nil
}

func (r *RedisInteractiveCache) ToRedisZ(ts domain.TopWithScore) redis.Z {
	return redis.Z{
		Score:  ts.Score,
		Member: ts.Member,
	}
}
