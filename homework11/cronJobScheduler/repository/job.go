package repository

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/repository/dao"
)

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next_time int64) error
	Stop(ctx context.Context, id int64) error
}

type PreemptCronJobRepository struct {
	dao dao.CronJobDAO
}

func NewPreemptCronJobRepository(dao dao.CronJobDAO) CronJobRepository {
	return &PreemptCronJobRepository{
		dao: dao,
	}
}

func (repo *PreemptCronJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	job, err := repo.dao.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	return repo.toDomain(job), nil
}

func (repo *PreemptCronJobRepository) Release(ctx context.Context, id int64) error {
	return repo.dao.Release(ctx, id)
}

func (repo *PreemptCronJobRepository) toEntity(job domain.Job) dao.Job {
	return dao.Job{
		Id:   job.Id,
		Name: job.Name,
		Cfg:  job.Cfg,
	}
}

func (repo *PreemptCronJobRepository) toDomain(job dao.Job) domain.Job {
	return domain.Job{
		Id:           job.Id,
		Name:         job.Name,
		Cfg:          job.Cfg,
		ExecutorName: job.ExecutorName,
		Cron:         job.Cron,
	}
}

func (repo *PreemptCronJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	return repo.dao.UpdateUtime(ctx, id)
}

func (repo *PreemptCronJobRepository) UpdateNextTime(ctx context.Context, id int64, next_time int64) error {
	return repo.dao.UpdateNextTime(ctx, id, next_time)
}

func (repo *PreemptCronJobRepository) Stop(ctx context.Context, id int64) error {
	return repo.dao.Stop(ctx, id)
}
