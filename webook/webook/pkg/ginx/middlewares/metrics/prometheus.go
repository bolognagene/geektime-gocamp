package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type Builder struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	// 在Prometheus的术语里，一个能够被抓取监控数据的endpoint叫做Instance，
	// 通常可以认为是一个进程。有着同样目的的Instance集合叫做Job
	InstanceID string
}

func NewBuilder(namespace string, subsystem string,
	name string, help string, instanceID string) *Builder {
	return &Builder{
		Namespace:  namespace,
		Subsystem:  subsystem,
		Name:       name,
		Help:       help,
		InstanceID: instanceID,
	}
}

func (b *Builder) Build() gin.HandlerFunc {
	// pattern 是指你命中的路由
	// 是指你的 HTTP 的 status
	// path /detail/1
	// 记录每个请求的响应时间
	labels := []string{"method", "pattern", "status"}
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      b.Name + "_resp_name",
		Help:      b.Help,
		ConstLabels: map[string]string{
			"instance_id": b.InstanceID,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		},
	}, labels)
	prometheus.MustRegister(summary)

	// 记录当前活跃的请求
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      b.Name + "_active_req",
		Help:      b.Help,
		ConstLabels: map[string]string{
			"instance_id": b.InstanceID,
		},
	})
	prometheus.MustRegister(gauge)

	return func(context *gin.Context) {
		start := time.Now()
		gauge.Inc()
		defer func() {
			duration := time.Since(start)
			gauge.Dec()
			pattern := context.FullPath()
			if pattern == "" {
				pattern = "unknown"
			}
			summary.WithLabelValues(context.Request.Method,
				pattern, strconv.Itoa(context.Writer.Status())).Observe(float64(duration.Milliseconds()))
		}()
		// 你最终就会执行到业务里面
		context.Next()
	}
}
