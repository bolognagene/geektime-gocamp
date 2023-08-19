package dao

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/domain"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

type UserProfileDao struct {
	db *gorm.DB
}

func NewUserProfileDAO(db *gorm.DB) *UserProfileDao {
	return &UserProfileDao{
		db: db,
	}
}

func (dao *UserProfileDao) FindByUserId(ctx context.Context, uid int64) (UserProfile, error) {
	// SELECT * FROM USER_PROFILE WHERE USER_ID = ?
	var up UserProfile
	err := dao.db.WithContext(ctx).Where("user_id = ?", uid).First(&up).Error
	//err := dao.db.WithContext(ctx).First(&u, "email = ?", email).Error
	return up, err
}

func (dao *UserProfileDao) Insert(ctx context.Context, up UserProfile) error {
	// 存毫秒数
	now := time.Now().UnixMilli()
	up.Utime = now
	up.Ctime = now
	err := dao.db.WithContext(ctx).Create(&up).Error
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			// 邮箱冲突
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (dao *UserProfileDao) Update(ctx context.Context, up domain.UserProfile) error {
	// 存毫秒数
	now := time.Now().UnixMilli()

	// 先找
	upp, err := dao.FindByUserId(ctx, up.UserId)
	if err == ErrUserNotFound {
		upp.Ctime = now
		upp.UserId = up.UserId
	} else if err != nil {
		return err
	}

	upp.NickName = up.NickName
	upp.Birthday = up.Birthday
	upp.Intro = up.Intro
	upp.Utime = now

	err = dao.db.WithContext(ctx).Save(&upp).Error
	if err != nil {
		return err
	}
	return nil
}

// UserProfile 直接对应数据库表user_profile结构
type UserProfile struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 对应user表的Id
	UserId   int64 `gorm:"unique"`
	NickName string
	Birthday string
	Intro    string
	// 创建时间，毫秒数
	Ctime int64
	// 更新时间，毫秒数
	Utime int64
}
