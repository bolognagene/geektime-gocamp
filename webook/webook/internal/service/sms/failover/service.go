package failover

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/sms"
	"sync/atomic"
)

// 连续超时多少个后就切换sms供应商
type TimeoutFailoverSMSService struct {
	svcs      []sms.Service
	cnt       int32 // 连续超时的个数
	idx       int32
	threshold int32 // 阈值 连续超时超过这个数字，就要切换
}

func NewTimeoutFailoverSMSService(svcs []sms.Service, threshold int32) *TimeoutFailoverSMSService {
	return &TimeoutFailoverSMSService{
		svcs:      svcs,
		threshold: threshold,
		idx:       0,
		cnt:       0,
	}
}

func (t TimeoutFailoverSMSService) Send(ctx context.Context, tpl string, args []sms.NamedArg, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)

	if cnt > t.threshold {
		// 这里要切换，新的下标，往后挪了一个
		newIdx := (idx + 1) % (int32)(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 我成功往后挪了一位
			atomic.StoreInt32(&t.cnt, 0)
		}
		// else 就是出现并发，别人换成功了，我要再load一遍
		idx = atomic.LoadInt32(&t.idx)
	}

	// 没有超过threshold, 就直接发送
	err := t.svcs[idx].Send(ctx, tpl, args, numbers...)
	switch err {
	case context.DeadlineExceeded:
		// cnt +1
		atomic.AddInt32(&t.cnt, 1)
		return err
	case nil:
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	default:
		// 不知道什么错误

		// 你可以考虑，换下一个，语义则是：
		// - 超时错误，可能是偶发的，我尽量再试试
		// - 非超时，我直接下一个
		return err
	}
}

func (t TimeoutFailoverSMSService) GetVendor() string {
	return t.svcs[atomic.LoadInt32(&t.idx)].GetVendor()
}
