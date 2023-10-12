package memory

// This service is for simulating Cloopen SMS service （容联云）
// Just for test
import (
	"context"
	"errors"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
)

type CloopenService struct {
}

func NewCloopenService() sms.Service {
	return &CloopenService{}
}

func (s *CloopenService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	fmt.Println(args)
	return nil
}

func (s *CloopenService) SendV1(ctx context.Context, tpl string, args []string, numbers ...string) error {
	fmt.Println(args)
	return errors.New("容联云短信服务发送失败")
}
