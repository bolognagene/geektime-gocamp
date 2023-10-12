package cloopen

import (
	"context"
	"errors"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
)

type FeeService struct {
	cnt  int32  // 该短信服务商发送短信计数
	name string // 服务商名字
}

func NewFeeService() sms.FeeService {
	return &FeeService{}
}

func (f *FeeService) Fee(ctx context.Context, args ...any) (float32, error) {
	// 容联云根据短信发送数量来计算每条的费用
	cnt, ok := args[0].(int32)

	if !ok {
		return 1.0, errors.New("输入参数类型不对")
	}

	var fee float32
	if cnt < 10417 {
		fee = 0.048
	} else if cnt < 22222 {
		fee = 0.045
	} else if cnt < 47619 {
		fee = 0.042
	} else if cnt < 125000 {
		fee = 0.040
	}

	return fee, nil
}

func (f *FeeService) GetName() string {
	return f.name
}

func (f *FeeService) GetCnt() *int32 {
	return &f.cnt
}
