// Package cloopen 容联云短信的实现
// SDK文档:https://doc.yuntongxun.com/pe/5f029a06a80948a1006e7760
package cloopen

import (
	"context"
	"fmt"
	mysms "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
	"github.com/ecodeclub/ekit/slice"
	"log"

	"github.com/cloopen/go-sms-sdk/cloopen"
)

type Service struct {
	client *cloopen.SMS
	appId  string
	vendor string
}

func NewService(c *cloopen.SMS, addId, vendor string) *Service {
	return &Service{
		client: c,
		appId:  addId,
		vendor: vendor,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []mysms.NamedArg, numbers ...string) error {
	input := &cloopen.SendRequest{
		// 应用的APPID
		AppId: s.appId,
		// 模版ID
		TemplateId: tplId,
		// 模版变量内容 非必填
		Datas: slice.Map[mysms.NamedArg, string](args, func(idx int, src mysms.NamedArg) string {
			return src.Val
		}),
	}

	for _, number := range numbers {
		// 手机号码
		input.To = number

		resp, err := s.client.Send(input)
		if err != nil {
			return err
		}

		if resp.StatusCode != "000000" {
			log.Printf("response code: %s, msg: %s \n", resp.StatusCode, resp.StatusMsg)
			fmt.Errorf("发送失败，code: %s, 原因：%s",
				resp.StatusCode, resp.StatusMsg)
		}
	}
	return nil
}

func (s *Service) GetVendor() string {
	return s.vendor
}
