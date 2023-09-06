package service

import (
	"context"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/service/sms"
	"math/rand"
)

const (
	defaultTplId = "1"
	loginTplId   = "1"
	orderTplId   = "2"
)

type SMSCodeService struct {
	repo   *repository.SMSCodeRepository
	smsSvc sms.Service
}

func NewSMSCodeService(repo *repository.SMSCodeRepository, smsSvc sms.Service) *SMSCodeService {
	return &SMSCodeService{repo: repo, smsSvc: smsSvc}
}

// Send 发验证码，我需要什么参数？
func (svc *SMSCodeService) Send(ctx context.Context,
	// 区别业务场景
	biz string,
	phone string) error {
	// 生成一个验证码
	code := svc.generateCode()
	// 塞进去 Redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		// 有问题
		return err
	}
	// 这前面成功了
	// 发送出去
	err = svc.smsSvc.Send(ctx, func(biz string) string {
		if biz == "login" {
			return loginTplId
		} else if biz == "order" {
			return orderTplId
		} else {
			return defaultTplId
		}
	}(biz), []string{code}, phone)
	//if err != nil {
	// 这个地方怎么办？
	// 这意味着，Redis 有这个验证码，但是不好意思，
	// 我能不能删掉这个验证码？
	// 你这个 err 可能是超时的 err，你都不知道，发出了没
	// 在这里重试
	// 要重试的话，初始化的时候，传入一个自己就会重试的 smsSvc
	//}
	return err
}

func (svc *SMSCodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *SMSCodeService) generateCode() string {
	// 六位数，num 在 0, 999999 之间，包含 0 和 999999
	num := rand.Intn(1000000)
	// 不够六位的，加上前导 0
	// 000001
	return fmt.Sprintf("%06d", num)
}
