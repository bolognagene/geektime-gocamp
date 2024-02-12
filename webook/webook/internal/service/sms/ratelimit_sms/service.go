package ratelimit_sms

import (
	"context"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/ratelimit"
)

var errLimited = fmt.Errorf("触发了限流")

type RatelimitSMSService struct {
	svc   sms.Service
	limit ratelimit.Limiter
}

func NewRatelimitSMSService(svc sms.Service, limit ratelimit.Limiter) *RatelimitSMSService {
	return &RatelimitSMSService{
		svc:   svc,
		limit: limit,
	}
}

// 装饰器模式
func (r RatelimitSMSService) Send(ctx context.Context, tpl string, args []sms.NamedArg, numbers ...string) error {
	limited, err := r.limit.Limit(ctx, fmt.Sprintf("sms:ratelimit:%s", r.svc.GetVendor()))
	if err != nil {
		return err
	}
	if limited {
		return errLimited
	}

	// 你这里加一些代码，新特性
	err = r.svc.Send(ctx, tpl, args, numbers...)
	// 你在这里也可以加一些代码，新特性
	return err
}

func (r RatelimitSMSService) GetVendor() string {
	return r.svc.GetVendor()
}
