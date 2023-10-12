package repository

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"time"
)

type SMSAsyncRepository interface {
	Create(ctx context.Context, sa domain.SMSAsync) error
	FindAll(ctx context.Context) ([]domain.SMSAsync, error)
	Remove(ctx context.Context, sa domain.SMSAsync) error
}

type GormSMSAsyncRepository struct {
	dao dao.SMSDAO
}

func NewSMSAsyncRepository(dao dao.SMSDAO) SMSAsyncRepository {
	return &GormSMSAsyncRepository{
		dao: dao,
	}
}

func (repo *GormSMSAsyncRepository) Create(ctx context.Context, sa domain.SMSAsync) error {
	return repo.dao.Insert(ctx, repo.domainToEntity(sa))
}

func (repo *GormSMSAsyncRepository) FindAll(ctx context.Context) ([]domain.SMSAsync, error) {
	smses, err := repo.dao.Find(ctx)
	if err != nil {
		return nil, err
	}

	var smsAsyncs []domain.SMSAsync
	for _, sms := range smses {
		smsAsyncs = append(smsAsyncs, repo.entityToDomain(sms))
	}

	return smsAsyncs, nil
}

func (repo *GormSMSAsyncRepository) Remove(ctx context.Context, sa domain.SMSAsync) error {
	return repo.dao.Delete(ctx, repo.domainToEntity(sa))
}

func (repo *GormSMSAsyncRepository) domainToEntity(sa domain.SMSAsync) dao.SMS {
	return dao.SMS{
		Id:      sa.Id,
		Biz:     sa.Biz,
		Args:    sa.Args,
		Numbers: sa.Numbers,
		Ctime:   sa.Ctime.UnixMilli(),
	}
}

func (repo *GormSMSAsyncRepository) entityToDomain(s dao.SMS) domain.SMSAsync {
	return domain.SMSAsync{
		Id:      s.Id,
		Biz:     s.Biz,
		Args:    s.Args,
		Numbers: s.Numbers,
		Ctime:   time.UnixMilli(s.Ctime),
	}
}
