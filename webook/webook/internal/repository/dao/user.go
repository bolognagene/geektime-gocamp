package dao

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicatePhone = errors.New("电话号码冲突")
	ErrUserNotFound       = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) Insert(ctx *gin.Context, u User) error {
	now := time.Now().UnixMilli()
	u.CTime = now
	u.UTime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	mysqlErr, ok := err.(*mysql.MySQLError)
	if ok {
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictsErrNo {
			return ErrUserDuplicatePhone
		}
	}

	return err

}

func (dao *UserDAO) FindByEmail(ctx *gin.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).Find(&u).Error
	return u, err
}
func (dao *UserDAO) FindByPhone(ctx *gin.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).Find(&u).Error
	return u, err
}

func (dao *UserDAO) UpdateProfile(ctx *gin.Context, u User) error {
	now := time.Now().UnixMilli()
	u.UTime = now
	return dao.db.WithContext(ctx).Model(&u).Updates(
		User{
			Nickname: u.Nickname,
			Birthday: u.Birthday,
			Intro:    u.Intro,
			UTime:    u.UTime,
		}).Error
}

func (dao *UserDAO) UpdatePassword(ctx *gin.Context, u User) error {
	now := time.Now().UnixMilli()
	u.UTime = now
	return dao.db.WithContext(ctx).Model(&u).Updates(User{
		Password: u.Password,
		UTime:    u.UTime,
	}).Error
}

func (dao *UserDAO) QueryProfile(ctx *gin.Context, u User) (User, error) {
	err := dao.db.WithContext(ctx).First(&u).Error

	return u, err
}

func (dao *UserDAO) FindById(ctx *gin.Context, id int64) (User, error) {
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

	// 创建时间，毫秒数
	CTime int64
	// 更新时间，毫秒数
	UTime int64
}
