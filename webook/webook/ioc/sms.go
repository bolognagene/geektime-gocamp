package ioc

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms/failover"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms/memory"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms/ratelimit_sms"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
	"time"
)

func InitSMSService(cmd redis.Cmdable) sms.Service {
	// 换内存，还是换别的
	//return memory.NewService("memory")

	// TimeoutFailoverSMSService
	memSvc := memory.NewService("memory")
	tencentSvc := memory.NewService("tencent")
	aliyunSvc := memory.NewService("aliyun")
	cloopenSvc := memory.NewService("cloopen")

	limiter := ratelimit.NewRedisSlidingWindowLimiter(cmd, time.Minute, 3000)

	memRateLimitSvc := ratelimit_sms.NewRatelimitSMSService(memSvc, limiter)
	tencentRateLimitSvc := ratelimit_sms.NewRatelimitSMSService(tencentSvc, limiter)
	aliyunRateLimitSvc := ratelimit_sms.NewRatelimitSMSService(aliyunSvc, limiter)
	cloopenRateLimitSvc := ratelimit_sms.NewRatelimitSMSService(cloopenSvc, limiter)

	svcs := []sms.Service{memRateLimitSvc, tencentRateLimitSvc, aliyunRateLimitSvc, cloopenRateLimitSvc}
	return failover.NewTimeoutFailoverSMSService(svcs, 5)

}
