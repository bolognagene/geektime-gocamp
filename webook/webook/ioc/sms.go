package ioc

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	// 换内存，还是换别的
	return memory.NewService()
}
