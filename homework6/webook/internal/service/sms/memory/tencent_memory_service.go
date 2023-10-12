package memory

// This service is for simulating Tencent SMS service
// Just for test
import (
	"context"
	"errors"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
)

type TencentService struct {
}

func NewTencentService() sms.Service {
	return &TencentService{}
}

func (s *TencentService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	fmt.Println(args)
	return nil
}

func (s *TencentService) SendV1(ctx context.Context, tpl string, args []string, numbers ...string) error {
	fmt.Println(args)
	return errors.New("腾讯云短信服务发送失败")
}
