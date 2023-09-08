package ioc

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/service/sms"
	sms_cloopen "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/service/sms/cloopen"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/service/sms/memory"
	"github.com/cloopen/go-sms-sdk/cloopen"
)

func InitMemorySMSService() sms.Service {
	// Memory
	return memory.NewService()
}

func InitCloopenSMSService(s *cloopen.SMS, applId string) sms.Service {
	// Cloopen
	return sms_cloopen.NewService(s, applId)
}
