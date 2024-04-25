package job

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// RankingJobAdapter 这里实现的是cron里的robfig的Job接口
type RankingJobAdapter struct {
	j Job
	l logger.Logger
	p prometheus.Summary
}

func NewRankingJobAdapter(j Job, l logger.Logger) *RankingJobAdapter {
	p := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "Golang",
		Subsystem: "webook",
		Name:      "cron_job",
		ConstLabels: map[string]string{
			"name": j.Name(),
		},
	})
	prometheus.MustRegister(p)
	return &RankingJobAdapter{
		j: j,
		l: l,
		p: p,
	}
}

func (r *RankingJobAdapter) Run() {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.p.Observe(float64(duration))
	}()
	err := r.j.Run()
	if err != nil {
		r.l.Error("运行任务失败", logger.Error(err),
			logger.String("job", r.j.Name()))
	}

}
