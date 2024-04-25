package redisx

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"net"
	"strconv"
	"time"
)

type PrometheusHook struct {
	vector *prometheus.SummaryVec
}

func NewPrometheusHook(opts prometheus.SummaryOpts) *PrometheusHook {
	vector := prometheus.NewSummaryVec(opts, []string{"cmd", "key_exist"})
	prometheus.MustRegister(vector)

	return &PrometheusHook{
		vector: vector,
	}
}

func (p *PrometheusHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 相当于这里什么都没做，类似于context.next()
		return next(ctx, network, addr)
	}
}

func (p *PrometheusHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		// 在Redis执行之前执行
		startTime := time.Now()
		var err error
		defer func() {
			duration := time.Since(startTime).Milliseconds()
			keyExist := err == redis.Nil
			p.vector.WithLabelValues(cmd.Name(), strconv.FormatBool(keyExist)).
				Observe(float64(duration))
		}()

		// 这个会最终发送命令到 redis 上
		err = next(ctx, cmd)

		return err
	}

}

func (p *PrometheusHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}

}
