package service

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"golang.org/x/sync/errgroup"
)

type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, bizId, uid, limit int64) error
	Unlike(ctx context.Context, biz string, bizId, uid, limit int64) error
	AddCollect(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	DeleteCollect(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error)
	TopLike(ctx context.Context, biz string, n, limit int64) ([]domain.TopWithScore, error)
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
}

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{
		repo: repo,
	}
}

func (svc *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return svc.repo.IncrReadCnt(ctx, biz, bizId)
}

func (svc *interactiveService) Like(ctx context.Context, biz string, bizId, uid, limit int64) error {
	return svc.repo.IncrLike(ctx, biz, bizId, uid, limit)
}

func (svc *interactiveService) Unlike(ctx context.Context, biz string, bizId, uid, limit int64) error {
	return svc.repo.DecrLike(ctx, biz, bizId, uid, limit)
}

func (svc *interactiveService) AddCollect(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	return svc.repo.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

func (svc *interactiveService) DeleteCollect(ctx context.Context, biz string, bizId int64, cid int64, uid int64) error {
	return svc.repo.DeleteCollectionItem(ctx, biz, bizId, cid, uid)
}

func (svc *interactiveService) Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error) {
	var (
		eg          errgroup.Group
		interactive domain.Interactive
		liked       bool
		collected   bool
		err         error
	)

	eg.Go(func() error {
		interactive, err = svc.repo.GetCnt(ctx, biz, bizId)
		return err
	})

	eg.Go(func() error {
		liked, err = svc.repo.Liked(ctx, biz, bizId, uid)
		return err
	})

	eg.Go(func() error {
		collected, err = svc.repo.Collected(ctx, biz, bizId, uid)
		return err
	})

	err = eg.Wait()
	if err != nil {
		return domain.Interactive{}, err
	}

	interactive.Liked = liked
	interactive.Collected = collected
	return interactive, err

}

func (svc *interactiveService) TopLike(ctx context.Context, biz string, n, limit int64) ([]domain.TopWithScore, error) {
	return svc.repo.GetTopLike(ctx, biz, n, limit)
}

func (svc *interactiveService) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	//TODO implement me
	panic("implement me")
}
