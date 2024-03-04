package repository

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	DeleteCollectionItem(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	GetCnt(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
	l     logger.Logger
}

func NewCachedInteractiveRepository(dao dao.InteractiveDAO,
	cache cache.InteractiveCache,
	l logger.Logger) InteractiveRepository {
	return &CachedInteractiveRepository{
		dao:   dao,
		cache: cache,
		l:     l,
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

func (repo *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	// 先插入点赞，然后更新点赞计数，更新缓存
	err := repo.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}

	// 这种做法，你需要在 repository 层面上维持住事务
	return repo.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
}

func (repo *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := repo.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}

	return repo.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
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

// 正常来说，参数必然不用指针：方法不要修改参数，通过返回值来修改参数
// 返回值就看情况。如果是指针实现了接口，那么就返回指针
// 如果返回值很大，你不想值传递引发复制问题，那么还是返回指针
// 返回结构体

// 最简原则：
// 1. 接收器永远用指针
// 2. 输入输出都用结构体
func (c *CachedInteractiveRepository) toDomain(intr dao.Interactive) domain.Interactive {
	return domain.Interactive{
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		ReadCnt:    intr.ReadCnt,
	}
}
