package aliyun

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
	"strconv"
	"strings"
)

type Service struct {
	client   *dysmsapi.Client
	signName string
	vendor   string
}

func NewService(c *dysmsapi.Client, signName, vendor string) *Service {
	return &Service{
		client:   c,
		signName: signName,
		vendor:   vendor,
	}
}

func (s *Service) SendV1(ctx context.Context, tplId string, args []string, numbers ...string) error {
	req := dysmsapi.CreateSendSmsRequest()
	req.Scheme = "https"
	// 阿里云多个手机号为字符串逗号间隔
	req.PhoneNumbers = strings.Join(numbers, ",")
	req.SignName = s.signName
	// 传的是 JSON
	argsMap := make(map[string]string, len(args))
	for idx, arg := range args {
		argsMap[strconv.Itoa(idx)] = arg
	}
	// 这意味着，你的模板必须是 你的短信验证码是{0}
	// 你的短信验证码是{code}
	bCode, err := json.Marshal(argsMap)
	if err != nil {
		return err
	}
	req.TemplateParam = string(bCode)
	req.TemplateCode = tplId

	var resp *dysmsapi.SendSmsResponse
	resp, err = s.client.SendSms(req)
	if err != nil {
		return err
	}

	if resp.Code != "OK" {
		return fmt.Errorf("发送失败，code: %s, 原因：%s",
			resp.Code, resp.Message)
	}
	return nil
}

func (s *Service) Send(ctx context.Context, tplId string, args []sms.NamedArg, numbers ...string) error {
	req := dysmsapi.CreateSendSmsRequest()
	req.Scheme = "https"
	// 阿里云多个手机号为字符串逗号间隔
	req.PhoneNumbers = strings.Join(numbers, ",")
	req.SignName = s.signName
	// 传的是 JSON
	argsMap := make(map[string]string, len(args))
	for _, arg := range args {
		argsMap[arg.Name] = arg.Val
	}
	// 这意味着，你的模板必须是 你的短信验证码是{0}
	// 你的短信验证码是{code}
	bCode, err := json.Marshal(argsMap)
	if err != nil {
		return err
	}
	req.TemplateParam = string(bCode)
	req.TemplateCode = tplId

	var resp *dysmsapi.SendSmsResponse
	resp, err = s.client.SendSms(req)
	if err != nil {
		return err
	}

	if resp.Code != "OK" {
		return fmt.Errorf("发送失败，code: %s, 原因：%s",
			resp.Code, resp.Message)
	}
	return nil
}

func (s *Service) GetVendor() string {
	return s.vendor
}
