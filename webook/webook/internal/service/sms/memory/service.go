package memory

import (
	"context"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
)

type Service struct {
	vendor string
}

func NewService(vendor string) *Service {
	return &Service{
		vendor: vendor,
	}
}

func (s *Service) Send(ctx context.Context, tpl string, args []sms.NamedArg, numbers ...string) error {
	fmt.Println(args)
	return nil
}

func (s *Service) GetVendor() string {
	return s.vendor
}
