package cloopen

import (
	"context"
	"github.com/cloopen/go-sms-sdk/cloopen"
	"testing"
)

func TestSender(t *testing.T) {
	accountSId := "2c94811c8a27cf2d018a2be335100178"
	authToken := "ad9b857f4aac47e78e8e0753102dc89f"
	appId := "2c94811c8a27cf2d018a2be3365c017f"
	number := "15827557348"

	cfg := cloopen.DefaultConfig().
		WithAPIAccount(accountSId).
		WithAPIToken(authToken)
	c := cloopen.NewJsonClient(cfg).SMS()

	s := NewService(c, appId)

	tests := []struct {
		name    string
		tplId   string
		data    []string
		numbers []string
		wantErr error
	}{
		{
			name:  "发送验证码",
			tplId: "1",
			data:  []string{"123456", "10"},
			// 改成你的手机号码
			numbers: []string{number},
		},
	}
	for _, tt := range tests {
		err := s.Send(context.Background(), tt.tplId, tt.data, tt.numbers...)
		if err != nil {
			t.Fatal(err)
		}
	}
}
