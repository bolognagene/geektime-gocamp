package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type SMSDAO interface {
	Insert(ctx context.Context, s SMS) error
	Find(ctx context.Context) ([]SMS, error)
	Delete(ctx context.Context, sms SMS) error
}

type GORMSMSDAO struct {
	db *gorm.DB
}

func NewSMSDAO(db *gorm.DB) SMSDAO {
	return &GORMSMSDAO{
		db: db,
	}
}

func (dao *GORMSMSDAO) Insert(ctx context.Context, s SMS) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	s.Ctime = now
	err := dao.db.WithContext(ctx).Create(&s).Error

	return err
}

func (dao *GORMSMSDAO) Find(ctx context.Context) ([]SMS, error) {
	var smss []SMS
	err := dao.db.WithContext(ctx).Find(&smss, nil).Error
	return smss, err
}

func (dao *GORMSMSDAO) Delete(ctx context.Context, sms SMS) error {
	err := dao.db.WithContext(ctx).Delete(&sms).Error
	return err
}

// SMS 直接对应数据库表结构
type SMS struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 业务代码
	Biz     string
	Args    string
	Numbers string
	Ctime   int64
}
