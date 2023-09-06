// Package cloopen 容联云短信的实现
// SDK文档:https://doc.yuntongxun.com/pe/5f029a06a80948a1006e7760

package cloopen

import (
	"context"
	"fmt"
	"github.com/cloopen/go-sms-sdk/cloopen"
	"log"
)

type Service struct {
	client *cloopen.SMS
	appId  string
}

func NewService(s *cloopen.SMS, applId string) *Service {
	return &Service{
		client: s,
		appId:  applId,
	}
}

func (s *Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {

	for _, phoneNumber := range numbers {
		input := &cloopen.SendRequest{
			AppId:      s.appId,
			To:         phoneNumber,
			TemplateId: tpl,
			Datas:      args,
		}

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
