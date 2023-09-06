package repository

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/cache/sms"
)

type SMSCodeRepository struct {
	cache sms.CodeCache
}

func NewSMSCodeRepository(cache sms.CodeCache) *SMSCodeRepository {
	return &SMSCodeRepository{cache: cache}
}

func (s *SMSCodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return s.cache.Set(ctx, biz, phone, code)
}

func (s *SMSCodeRepository) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return s.cache.Verify(ctx, biz, phone, inputCode)
}
