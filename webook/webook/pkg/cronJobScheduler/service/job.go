package service

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"time"
)

type CronJobService interface {
	// Preempt 抢占
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, job domain.Job) error
}

type PreemptCronJobService struct {
	repo            repository.PreemptCronJobRepository
	refreshInterval time.Duration
	l               logger.Logger
}

func NewPreemptCronJobService(repo repository.PreemptCronJobRepository,
	refreshInterval time.Duration, l logger.Logger) CronJobService {
	return &PreemptCronJobService{
		repo:            repo,
		refreshInterval: refreshInterval,
		l:               l,
	}
}

func (svc *PreemptCronJobService) ResetNextTime(ctx context.Context, job domain.Job) error {
	nt := job.NextTime()
	if nt.IsZero() {
		// 没有下一次
		return nil
	}
	return svc.repo.UpdateNextTime(ctx, job.Id, nt.UnixMilli())
}

func (svc *PreemptCronJobService) Preempt(ctx context.Context) (domain.Job, error) {
	job, err := svc.repo.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}

	// 续约
	ticker := time.NewTicker(svc.refreshInterval)
	go func() {
		for range ticker.C {
			svc.refresh(job.Id)
		}
	}()

	job.CancelFunc = func() error {
		// 自己在这里释放掉
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return svc.repo.Release(ctx, job.Id)
	}

	return domain.Job{
		Cfg: job.Cfg,
	}, nil
}

func (svc *PreemptCronJobService) refresh(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := svc.repo.UpdateUtime(ctx, id)
	if err != nil {
		svc.l.Error("续约失败",
			logger.Error(err),
			logger.Int64("jid", id))
	}
	return err
}
