package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type CronJobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next_time int64) error
}

type GORMCronJobDAO struct {
	db *gorm.DB
}

func NewGORMCronJobDAO(db *gorm.DB) CronJobDAO {
	return &GORMCronJobDAO{
		db: db,
	}
}

func (dao *GORMCronJobDAO) Release(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	var job Job
	err := dao.db.WithContext(ctx).Model(&Job{}).
		Where("status = ?", jobStatusRunning).
		First(&job).Error
	if err != nil {
		return err
	}

	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ? AND version = ?",
		id, job.Version).Updates(map[string]any{
		"status":  jobStatusWaiting,
		"utime":   now,
		"version": job.Version + 1,
	}).Error
}

func (dao *GORMCronJobDAO) Preempt(ctx context.Context) (Job, error) {
	// 高并发情况下，大部分都是陪太子读书
	// 100 个 goroutine
	// 要转几次？ 所有 goroutine 执行的循环次数加在一起是
	// 1+2+3+4 +5 + ... + 99 + 100
	// 特定一个 goroutine，最差情况下，要循环一百次
	for {
		now := time.Now().UnixMilli()
		var job Job
		// 分布式任务调度系统
		// 1. 一次拉一批，我一次性取出 100 条来，然后，我随机从某一条开始，向后开始抢占
		// 2. 我搞个随机偏移量，0-100 生成一个随机偏移量。兜底：第一轮没查到，偏移量回归到 0
		// 3. 我搞一个 id 取余分配，status = ? AND next_time <=? AND id%10 = ? 兜底：不加余数条件，取next_time 最老的
		err := dao.db.WithContext(ctx).Model(&Job{}).
			Where("status = ? AND next_time <= ?", jobStatusWaiting, now).
			First(&job).Error
		if err != nil {
			return Job{}, err
		}

		// 找到了后要抢占，将状态置为jobStatusRunning
		// 两个 goroutine 都拿到 id =1 的数据
		// 能不能用 utime?
		// 乐观锁，CAS 操作，compare AND Swap
		// 有一个很常见的面试刷亮点：就是用乐观锁取代 FOR UPDATE
		// 面试套路（性能优化）：曾经用了 FOR UPDATE =>性能差，还会有死锁 => 我优化成了乐观锁
		res := dao.db.WithContext(ctx).Model(&Job{}).Where("id = ? AND version = ?",
			job.Id, job.Version).Updates(map[string]any{
			"status":  jobStatusRunning,
			"utime":   now,
			"version": job.Version + 1,
		})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			// 抢占失败，继续下一轮
			continue
		}
		return job, nil
	}

}

func (dao *GORMCronJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ? AND status = ?", id, jobStatusRunning).
		Updates(map[string]any{
			"utime": now,
		}).Error
}

func (dao *GORMCronJobDAO) UpdateNextTime(ctx context.Context, id int64, next_time int64) error {
	return dao.db.WithContext(ctx).Model(&Job{}).
		Where("id = ? AND status = ?", id, jobStatusRunning).
		Updates(map[string]any{
			"next_time": next_time,
		}).Error
}

type Job struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 比如说 ranking
	Name string `gorm:"unique"`

	Cfg string
	// 第一个问题：哪些任务可以抢？哪些任务已经被人占着？哪些任务永远不会被运行
	// 用状态来标记
	Status int
	// 另外一个问题，定时任务，我怎么知道，已经到时间了呢？
	// NextTime 下一次被调度的时间
	// next_time <= now 这样一个查询条件
	// and status = 0
	// 要建立索引
	// 更加好的应该是 next_time 和 status 的联合索引
	NextTime int64 `gorm:"index"`

	Version int
	// 创建时间，毫秒数
	Ctime int64
	// 更新时间，毫秒数
	Utime int64
}

const (
	// 可以被抢占
	jobStatusWaiting = iota
	// 已经被抢占
	jobStatusRunning
	// 暂停调度
	jobStatusPaused
)
