package service

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/cronJobScheduler/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
)

type Executor interface {
	Name() string
	Exec(ctx context.Context, job domain.Job) error
	Register(job domain.Job, fn func(ctx context.Context, job domain.Job) error)
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, job domain.Job) error
	l     logger.Logger
}

func NewLocalFuncExecutor(l logger.Logger) Executor {
	return &LocalFuncExecutor{
		l:     l,
		funcs: make(map[string]func(ctx context.Context, job domain.Job) error),
	}
}

func (e *LocalFuncExecutor) Register(job domain.Job, fn func(ctx context.Context, job domain.Job) error) {
	e.funcs[job.Name] = fn
}

func (e *LocalFuncExecutor) Name() string {
	return "localfunc"
}

func (e *LocalFuncExecutor) Exec(ctx context.Context, job domain.Job) error {
	fn, ok := e.funcs[job.Name]
	if !ok {
		// DEBUG 的时候最好中断
		// 线上就继续
		e.l.Error("执行函数未注册",
			logger.String("executor_name", job.ExecutorName),
			logger.String("Job_name", job.Name))
	}
	return fn(ctx, job)
}
