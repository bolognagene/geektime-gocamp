package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicate = errors.New("数据冲突")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

type UserDAO interface {
	Insert(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindById(ctx context.Context, id int64) (User, error)
	FindByWechat(ctx context.Context, openId string) (User, error)
	UpdateProfile(ctx context.Context, u User) error
	UpdatePassword(ctx context.Context, u User) error
}

type GORMUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{
		db: db,
	}
}

func (dao *GORMUserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	mysqlErr, ok := err.(*mysql.MySQLError)
	if ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			return ErrUserDuplicate
		}
	}

	return err

}

func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).Find(&u).Error
	return u, err
}
func (dao *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).Find(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindByWechat(ctx context.Context, openId string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("wechat_open_id = ?", openId).Find(&u).Error
	return u, err
}

func (dao *GORMUserDAO) UpdateProfile(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Utime = now
	return dao.db.WithContext(ctx).Model(&u).Updates(
		User{
			Nickname: u.Nickname,
			Birthday: u.Birthday,
			Intro:    u.Intro,
			Utime:    u.Utime,
		}).Error
}

func (dao *GORMUserDAO) UpdatePassword(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Utime = now
	return dao.db.WithContext(ctx).Model(&u).Updates(User{
		Password: u.Password,
		Utime:    u.Utime,
	}).Error
}

/*func (dao *GORMUserDAO) QueryProfile(ctx context.Context, u User) (User, error) {
	err := dao.db.WithContext(ctx).First(&u).Error

	return u, err
}*/

func (dao *GORMUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("Id = ?", id).First(&u).Error

	return u, err
}

// User 直接对应数据库表结构
// 有些人叫做 entity，有些人叫做 model，有些人叫做 PO(persistent object)
type User struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	Email    string
	Password string
	Nickname string
	Birthday string
	Intro    string
	Phone    string `gorm:"unique"`
	// 索引的最左匹配原则：
	// 假如索引在 <A, B, C> 建好了
	// A, AB, ABC 都能用
	// WHERE A =?
	// WHERE A = ? AND B =?    WHERE B = ? AND A =?
	// WHERE A = ? AND B = ? AND C = ?  ABC 的顺序随便换
	// WHERE 里面带了 ABC，可以用
	// WHERE 里面，没有 A，就不能用

	// 如果要创建联合索引，<unionid, openid>，用 openid 查询的时候不会走索引
	// <openid, unionid> 用 unionid 查询的时候，不会走索引
	// 微信的字段
	WechatUnionID sql.NullString
	WechatOpenID  sql.NullString `gorm:"unique"`

	// 创建时间，毫秒数
	Ctime int64
	// 更新时间，毫秒数
	Utime int64
}
