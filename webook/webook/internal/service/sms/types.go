package sms

import "context"

type Service interface {
	Send(ctx context.Context, tpl string, args []NamedArg, numbers ...string) error
	//SendV1(ctx context.Context, tpl string, args []string, numbers ...string) error
	// 调用者需要知道实现者需要什么类型的参数，是 []string，还是 map[string]string
	//SendV2(ctx context.Context, tpl string, args any, numbers ...string) error
	//SendV3(ctx context.Context, tpl string, args T, numbers ...string) error
	GetVendor() string
}

type NamedArg struct {
	Val  string
	Name string
}
