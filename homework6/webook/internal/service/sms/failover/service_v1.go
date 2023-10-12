package failover

import (
	"context"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms/async"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/ratelimit"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

var errLimited = fmt.Errorf("触发了限流")

// ExtensiveFailoverSMSService
// 这个failover策略考虑所有的超时、error、以及响应时间超过threshold就计数+1
// 如果连续错误超过阈值则切换到下一个服务商
type ExtensiveFailoverSMSService struct {
	//svcs         []sms.Service
	svcs         []async.SMSService
	feeSvcs      []sms.FeeService  // 两个service切片的下标一定要对上
	limiter      ratelimit.Limiter // 限流pkg
	repo         repository.SMSAsyncRepository
	idx          int32
	cnt          int32 // 连续超时的个数
	threshold    int32 // 阈值 连续超时超过这个数字，就要切换
	resThreshold int32 //连续超时超过这个数字(秒级别)，就要切换
}

func NewExtensiveFailoverSMSService(svcs []async.SMSService, feeSvcs []sms.FeeService,
	limiter ratelimit.Limiter, repo repository.SMSAsyncRepository,
	threshold int32, resThreshold int32) sms.Service {
	return &ExtensiveFailoverSMSService{
		svcs:         svcs,
		feeSvcs:      feeSvcs,
		limiter:      limiter,
		repo:         repo,
		threshold:    threshold,
		resThreshold: resThreshold,
	}
}

func (s *ExtensiveFailoverSMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	//先看限流
	limited, err := s.limiter.Limit(ctx, "sms:cnt")
	if err != nil {
		// 系统错误
		// 可以限流：保守策略，你的下游很坑的时候，
		// 可以不限：你的下游很强，业务可用性要求很高，尽量容错策略
		// 包一下这个错误

		// 这里可以做一个策略，比如阿里、腾讯这些大厂的服务可以放过去，其他的则限流
		return fmt.Errorf("短信服务判断是否限流出现问题，%w", err)
	}
	if limited {
		// 转异步
		s.repo.Create(ctx, domain.SMSAsync{
			Biz:     biz,
			Args:    strings.Join(args, "###"),
			Numbers: strings.Join(numbers, ";"),
			Ctime:   time.Now(),
		})
		return errLimited
	}

	idx := atomic.LoadInt32(&s.idx)
	cnt := atomic.LoadInt32(&s.cnt)
	if cnt > s.threshold {
		// 这里要切换，新的下标，往后挪了一个
		newIdx, err := s.SwitchNext(ctx, idx)
		if err != nil {
			return err
		}
		if atomic.CompareAndSwapInt32(&s.idx, idx, newIdx) {
			// 我成功往后挪了一位
			atomic.StoreInt32(&s.cnt, 0)
		}
		// else 就是出现并发，别人换成功了
		// 这个时候只需要重新Load idx就可以了
		idx = atomic.LoadInt32(&s.idx)
	}

	var svc async.SMSService = s.svcs[idx]
	t1 := time.Now()
	err = svc.Send(ctx, biz, args, numbers...)
	t2 := time.Now()
	switch err {
	case context.DeadlineExceeded:
		atomic.AddInt32(&s.cnt, 1)
		//转异步
		return err
	case nil:
		// 该短信服务商cnt要加一
		atomic.AddInt32(s.feeSvcs[idx].GetCnt(), 1)
		// 响应时间超过设定的时长也需要加一
		if t2.Sub(t1) > time.Duration(s.resThreshold)*time.Second {
			atomic.AddInt32(&s.cnt, 1)
			//但是这里不用转异步，因为已经发出去了
		} else {
			// 你的连续状态被打断了
			atomic.StoreInt32(&s.cnt, 0)
			// 发送异步存储的短信
			svc.StartAsync(ctx)
		}

		return nil
	default:
		// 不知道什么错误
		// 你可以考虑，换下一个，语义则是：
		// - 超时错误，可能是偶发的，我尽量再试试
		// - 非超时，我直接下一个
		atomic.AddInt32(&s.cnt, 1)
		//转异步
		return err
	}
}

type idxAndFee struct {
	idx int32
	fee float32
}

// SwitchNext
// 根据费用切换到下一个
func (s *ExtensiveFailoverSMSService) SwitchNext(ctx context.Context, idx int32) (int32, error) {
	// 待写
	//return (idx + 1) % int32(len(f.svcs)), nil
	var i int32
	idxAndFees := make([]idxAndFee, len(s.feeSvcs))

	for i = 0; i < int32(len(s.feeSvcs)); i++ {

		fee, err := s.feeSvcs[i].Fee(ctx, s.feeSvcs[i].GetCnt())
		if err != nil {
			fee = 1.0 //自动排到最后
		}

		idxAndFees[i] = idxAndFee{
			idx: i,
			fee: fee,
		}
	}

	sort.Slice(idxAndFees, func(i, j int) bool {
		return idxAndFees[i].fee < idxAndFees[j].fee
	})

	for i = 0; i < int32(len(idxAndFees)); i++ {
		if idxAndFees[i].idx == idx {
			nextIdx := (i + 1) % int32(len(idxAndFees))
			return idxAndFees[nextIdx].idx, nil
		}
	}

	return (idx + 1) % int32(len(s.svcs)), nil

}
