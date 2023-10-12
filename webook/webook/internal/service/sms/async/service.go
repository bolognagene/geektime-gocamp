package async

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
	"strings"
	"time"
)

type SMSService struct {
	svc      sms.Service
	repo     repository.SMSAsyncRepository
	retryMax int // 重试
}

func NewSMSService(svc sms.Service, repo repository.SMSAsyncRepository, retryMax int) *SMSService {
	return &SMSService{
		svc:      svc,
		repo:     repo,
		retryMax: retryMax,
	}
}

func (s *SMSService) StartAsync(ctx context.Context) {
	go func() {
		smsAsyncs, err := s.repo.FindAll(ctx)
		if err != nil {
			return
		}
		for _, smsAsync := range smsAsyncs {
			// 在这里发送，并且控制重试
			var nums []string
			numbers := strings.Split(smsAsync.Numbers, ";")
			for _, number := range numbers {
				nums = append(nums, number)
			}

			cnt := 0
			for cnt < s.retryMax {
				err = s.svc.Send(ctx, smsAsync.Biz, strings.Split(smsAsync.Args, "###"),
					nums...)
				if err == nil {
					// 发送成功删除记录
					s.repo.Remove(ctx, smsAsync)
					break
				}
				cnt++
			}
		}
	}()
}

func (s *SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {

	var err error
	for cnt := 0; cnt < s.retryMax; cnt++ {
		err = s.svc.Send(ctx, biz, args, numbers...)
		if err == nil {
			return nil
		}
	}

	// 发送失败，存储请求
	s.repo.Create(ctx, domain.SMSAsync{
		Biz:     biz,
		Args:    strings.Join(args, "###"),
		Numbers: strings.Join(numbers, ";"),
		Ctime:   time.Now(),
	})

	return err

}
