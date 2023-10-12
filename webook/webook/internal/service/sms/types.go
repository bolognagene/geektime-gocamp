package sms

import "context"

type Service interface {
	// Send biz 很含糊的业务
	Send(ctx context.Context, biz string, args []string, numbers ...string) error
	//SendV1(ctx context.Context, tpl string, args []NamedArg, numbers ...string) error
	// 调用者需要知道实现者需要什么类型的参数，是 []string，还是 map[string]string
	//SendV2(ctx context.Context, tpl string, args any, numbers ...string) error
	//SendVV3(ctx context.Context, tpl string, args T, numbers ...string) error
	// 计算单条短信的费用
}

type NamedArg struct {
	Val  string
	Name string
}

type FeeService interface {
	Fee(ctx context.Context, args ...any) (float32, error)
	GetName() string
	GetCnt() *int32
}
