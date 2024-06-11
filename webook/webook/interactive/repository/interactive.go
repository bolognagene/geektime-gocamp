package repository

import (
	"context"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interacitve/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interacitve/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interacitve/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	BatchIncrReadCnt(ctx context.Context, biz string, bizIds []int64) error
	IncrLike(ctx context.Context, biz string, bizId, uid, limit int64) error
	DecrLike(ctx context.Context, biz string, bizId, uid, limit int64) error
	GetTopLike(ctx context.Context, biz string, n int64, limit int64) ([]domain.TopWithScore, error)
	AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	DeleteCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	GetCnt(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error)
}

type CachedInteractiveRepository struct {
	dao    dao.InteractiveDAO
	cache  cache.InteractiveCache
	client redis.Cmdable
	l      logger.Logger
}

func NewCachedInteractiveRepository(dao dao.InteractiveDAO,
	cache cache.InteractiveCache,
	client redis.Cmdable,
	l logger.Logger) InteractiveRepository {
	return &CachedInteractiveRepository{
		dao:    dao,
		cache:  cache,
		client: client,
		l:      l,
	}
}

func (repo *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	// 要考虑缓存方案了
	// 这两个操作能不能换顺序？ —— 不能
	err := repo.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}

	go func() {
		repo.cache.IncrReadCntIfPresent(ctx, biz, bizId)
	}()
	return err

	//return repo.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractiveRepository) BatchIncrReadCnt(ctx context.Context, biz string, bizIds []int64) error {
	// 要考虑缓存方案了
	// 这两个操作能不能换顺序？ —— 不能
	err := repo.dao.BatchIncrReadCnt(ctx, biz, bizIds)
	if err != nil {
		return err
	}

	go func() {
		repo.cache.BatchIncrReadCntIfPresent(ctx, biz, bizIds)
	}()
	return err
}

func (repo *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, bizId, uid, limit int64) error {
	// 先插入点赞，然后更新点赞计数，更新缓存
	err := repo.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}

	// 这种做法，你需要在 repository 层面上维持住事务
	go func() {
		ret, err1 := repo.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
		if err1 != nil {
			repo.l.Debug("增加点赞计数失败", logger.String("biz", biz),
				logger.Int64("bizId", bizId), logger.Error(err1))
		}

		fmt.Sprintf("ret is %d", ret)
	}()

	// 增加Toplike里的计数
	go func() {
		ret2, err2 := repo.cache.IncrTopLike(ctx, biz, bizId, limit)
		if err2 != nil {
			repo.l.Debug("增加TopLike计数失败，可能是该文章不在TopLike中", logger.String("biz", biz),
				logger.Int64("bizId", bizId), logger.Error(err2))
		}

		fmt.Sprintf("ret is %d", ret2)
	}()

	return nil
}

func (repo *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, bizId, uid, limit int64) error {
	err := repo.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}

	go func() {
		err1 := repo.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
		if err1 != nil {
			repo.l.Debug("取消点赞计数失败", logger.String("biz", biz),
				logger.Int64("bizId", bizId), logger.Error(err1))
		}
	}()

	// 减少Toplike里的计数
	go func() {
		err2 := repo.cache.DecrTopLike(ctx, biz, bizId, limit)
		if err2 != nil {
			repo.l.Debug("减少TopLike计数失败，可能是该文章不在TopLike中", logger.String("biz", biz),
				logger.Int64("bizId", bizId), logger.Error(err2))
		}
	}()

	return nil
}

func (repo *CachedInteractiveRepository) GetTopLike(ctx context.Context, biz string, n int64, limit int64) ([]domain.TopWithScore, error) {
	// 从缓存里读取TopLikeN
	data, err := repo.cache.GetTopLike(ctx, biz, n)

	if err == nil && data != nil {
		return data, nil
	}

	// 否则从数据库里查询
	intrs, err := repo.dao.GetTopLike(ctx, biz, limit)
	if err != nil {
		return nil, err
	}

	data = make([]domain.TopWithScore, len(intrs))
	for i, z := range intrs {
		data[i] = repo.ToTopWithScore(z)
	}

	go func() {
		// 回写到缓存中
		err1 := repo.cache.SetTopLike(ctx, biz, data)
		if err1 != nil {
			repo.l.Debug("设置toplike缓存失败", logger.String("biz", biz),
				logger.Error(err1))
		}
	}()

	if int64(len(data)) > n {
		return data[:n], nil
	}
	return data, nil

}

func (repo *CachedInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	// 这个地方，你要不要考虑缓存收藏夹？
	// 以及收藏夹里面的内容
	// 用户会频繁访问他的收藏夹，那么你就应该缓存，不然你就不需要
	// 一个东西要不要缓存，你就看用户会不会频繁访问（反复访问）
	err := repo.dao.InsertCollectionInfo(ctx, biz, bizId, cid, uid)
	if err != nil {
		return err
	}

	return repo.cache.IncrCollectCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractiveRepository) DeleteCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	err := repo.dao.DeleteCollectionInfo(ctx, biz, bizId, cid, uid)
	if err != nil {
		return err
	}

	return repo.cache.DecrCollectCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractiveRepository) GetCnt(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	// 要从缓存拿出来阅读数，点赞数和收藏数
	interactive, err := repo.cache.GetCnt(ctx, biz, bizId)
	if err == nil &&
		(interactive.CollectCnt != 0 ||
			interactive.LikeCnt != 0 ||
			interactive.ReadCnt != 0) {
		return interactive, nil
	}

	// 但不是所有的结构体都是可比较的
	//if intr == (domain.Interactive{}) {
	//
	//}
	// 在这里查询数据库
	daointer, err := repo.dao.GetInteractive(ctx, biz, bizId)
	if err != nil {
		return interactive, err
	}

	interactive = repo.toDomain(daointer)

	//回写缓存
	go func() {
		err1 := repo.cache.SetCnt(ctx, biz, bizId, interactive)
		if err1 != nil {
			//记录日志
			repo.l.Debug("repo.interactive回写缓存失败", logger.String("biz", biz),
				logger.Int64("bizId", bizId), logger.Error(err1))
		}
	}()
	return interactive, nil
}

func (repo *CachedInteractiveRepository) Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := repo.dao.GetLikeInfo(ctx, biz, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		// 你要吞掉
		return false, nil
	default:
		return false, err
	}

}

func (repo *CachedInteractiveRepository) Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := repo.dao.GetCollectionInfo(ctx, biz, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		// 你要吞掉
		return false, nil
	default:
		return false, err
	}

}

func (repo *CachedInteractiveRepository) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	interactives := make(map[int64]domain.Interactive)
	intrs, err := repo.dao.GetByIds(ctx, biz, bizIds)
	if err != nil {
		return interactives, err
	}

	for id, intr := range intrs {
		interactives[id] = repo.toDomain(intr)
	}

	return interactives, nil
}

// 正常来说，参数必然不用指针：方法不要修改参数，通过返回值来修改参数
// 返回值就看情况。如果是指针实现了接口，那么就返回指针
// 如果返回值很大，你不想值传递引发复制问题，那么还是返回指针
// 返回结构体

// 最简原则：
// 1. 接收器永远用指针
// 2. 输入输出都用结构体
func (repo *CachedInteractiveRepository) toDomain(intr dao.Interactive) domain.Interactive {
	return domain.Interactive{
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		ReadCnt:    intr.ReadCnt,
	}
}

func (repo *CachedInteractiveRepository) ToTopWithScore(intr dao.Interactive) domain.TopWithScore {
	return domain.TopWithScore{
		Score:  float64(intr.LikeCnt),
		Member: intr.BizId,
	}
}
