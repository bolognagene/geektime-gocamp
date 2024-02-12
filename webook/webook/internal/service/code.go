package service

import (
	"context"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
	"github.com/spf13/viper"
	"go.uber.org/atomic"
	"math/rand"
)

// const codeTplId = "1877556"
// 测试Viper的监听功能
var CodeTplId atomic.String = atomic.String{}

var (
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
)

type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type codeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service // 因为这里的sms.Service是一个接口，类似指针，所以直接sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	// viper读取
	codeTemplateId := viper.GetString("TplId.code")
	if codeTemplateId == "" {
		codeTemplateId = "1877550"
	}
	CodeTplId.Store(codeTemplateId)
	return &codeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *codeService) Send(ctx context.Context, biz, phone string) error {
	// 生成一个验证码
	code := svc.generateCode()
	// 塞进去 Redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送出去
	codeArg := sms.NamedArg{
		Name: "Code",
		Val:  code,
	}
	return svc.smsSvc.Send(ctx, CodeTplId.Load(), []sms.NamedArg{codeArg}, phone)
}

func (svc *codeService) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, code)
}

func (svc *codeService) generateCode() string {
	num := rand.Intn(1000000)
	return fmt.Sprintf("%06d", num)
}
