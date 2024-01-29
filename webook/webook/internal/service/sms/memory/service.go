package memory

import (
	"context"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

// Send 模拟发送验证码， 仅用于测试
func (s Service) Send(ctx context.Context, tpl string, args []sms.NamedArg, numbers ...string) error {
	fmt.Println(args)
	return nil
}
