package sms

import "context"

type CodeCache interface {
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
	Key(biz, phone string) string
	Set(ctx context.Context, biz, phone, code string) error
}
