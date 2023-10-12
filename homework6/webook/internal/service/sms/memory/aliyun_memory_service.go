package memory

// This service is for simulating Aliyun SMS service
// Just for test
import (
	"context"
	"errors"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
)

type AliyunService struct {
}

func NewAliyunService() sms.Service {
	return &AliyunService{}
}

func (s *AliyunService) SendV1(ctx context.Context, tpl string, args []string, numbers ...string) error {
	fmt.Println(args)
	return nil
}

func (s *AliyunService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	fmt.Println(args)
	return errors.New("阿里云短信服务发送失败")
}
