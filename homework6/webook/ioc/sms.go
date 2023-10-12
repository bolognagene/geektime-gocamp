package ioc

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms/async"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms/failover"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms/memory"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
	"time"
)

// 这是实际的函数
/*func InitSMSService(cmd redis.Cmdable, repo repository.SMSAsyncRepository) sms.Service {
	// 换内存，还是换别的
	//svc := ratelimit.NewRatelimitSMSService(memory.NewService(),
	//	limiter.NewRedisSlidingWindowLimiter(cmd, time.Second, 100))
	//return retryable.NewService(svc, 3)
	//return memory.NewService()

	// aliyun服务
	aliClient, err := dysmsapi.NewClient()
	if err != nil {
		panic("阿里云短信服务初始化失败")
	}
	aliService := aliyunv1.NewService(aliClient, "aliyunSignName")
	asyncAliService := async.NewSMSService(aliService, repo, 3)
	aliFeeService := aliyunv1.NewFeeService()

	// tencent服务
	tpc := profile.NewClientProfile()
	credential := common.NewCredential(
		// os.Getenv("TENCENTCLOUD_SECRET_ID"),
		// os.Getenv("TENCENTCLOUD_SECRET_KEY"),
		"SecretId",
		"SecretKey",
	)
	tencentClient, err := v20210111.NewClient(credential, "China", tpc)
	if err != nil {
		panic("腾讯云短信服务初始化失败")
	}
	tencentService := tencent.NewService("tencentAppId", "tencentSignName", tencentClient)
	asyncTencentService := async.NewSMSService(tencentService, repo, 3)
	tencentFeeService := tencent.NewFeeService()

	// 容联云服务
	cfg := cloopen.DefaultConfig().WithAPIAccount("cloopenApiAccount").WithAPIToken("cloopenApiToken")
	cloopenClient := cloopen.NewJsonClient(cfg).SMS()
	cloopenService := myCloopen.NewService(cloopenClient, "cloopenApiId")
	asyncCloopenService := async.NewSMSService(cloopenService, repo, 3)
	cloopenFeeService := myCloopen.NewFeeService()

	//限流服务
	limiter := ratelimit.NewRedisSlidingWindowLimiter(cmd, 1*time.Second, 3000)

	//组装failoverService
	var smses []sms.Service
	var fees []sms.FeeService
	smses = append(smses, asyncAliService, asyncTencentService, asyncCloopenService)
	fees = append(fees, aliFeeService, tencentFeeService, cloopenFeeService)
	return failover.NewExtensiveFailoverSMSService(smses, fees, limiter, repo, 5, 1)
}*/

// 由于我的阿里云、腾讯云启动不成功（没有申请相关的数据）因此这里用memory来模拟阿里、腾讯和容联云的短信服务
func InitSMSService(cmd redis.Cmdable, repo repository.SMSAsyncRepository) sms.Service {
	// 换内存，还是换别的
	//svc := ratelimit.NewRatelimitSMSService(memory.NewService(),
	//	limiter.NewRedisSlidingWindowLimiter(cmd, time.Second, 100))
	//return retryable.NewService(svc, 3)
	//return memory.NewService()

	// aliyun服务
	aliService := memory.NewAliyunService()
	asyncAliService := async.NewSMSService(aliService, repo, 3)
	aliFeeService := memory.NewAliyunFeeService()

	// tencent服务
	tencentService := memory.NewTencentService()
	asyncTencentService := async.NewSMSService(tencentService, repo, 3)
	tencentFeeService := memory.NewTencentFeeService()

	// 容联云服务
	cloopenService := memory.NewCloopenService()
	asyncCloopenService := async.NewSMSService(cloopenService, repo, 3)
	cloopenFeeService := memory.NewCloopenFeeService()

	//限流服务
	limiter := ratelimit.NewRedisSlidingWindowLimiter(cmd, 1*time.Second, 3000)

	//组装failoverService
	var smses []async.SMSService
	var fees []sms.FeeService
	smses = append(smses, *asyncAliService, *asyncTencentService, *asyncCloopenService)
	fees = append(fees, aliFeeService, tencentFeeService, cloopenFeeService)
	return failover.NewExtensiveFailoverSMSService(smses, fees, limiter, repo, 2, 1)
}
