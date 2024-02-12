package repository

import (
	"database/sql"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"github.com/gin-gonic/gin"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository interface {
	Create(ctx *gin.Context, u domain.User) error
	FindByEmail(ctx *gin.Context, u domain.User) (domain.User, error)
	FindByPhone(ctx *gin.Context, u domain.User) (domain.User, error)
	FindById(ctx *gin.Context, id int64) (domain.User, error)
	FindByWechat(ctx *gin.Context, openId string) (domain.User, error)
	UpdateProfile(ctx *gin.Context, u domain.User) error
	UpdatePassword(ctx *gin.Context, u domain.User) error
}

type CachedUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CachedUserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *CachedUserRepository) Create(ctx *gin.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Phone:    u.Phone,
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *CachedUserRepository) FindByEmail(ctx *gin.Context, u domain.User) (domain.User, error) {
	user, err := r.dao.FindByEmail(ctx, u.Email)
	if err != nil {
		return domain.User{}, err
	}

	return r.entityToDomain(user), nil
}

func (r *CachedUserRepository) FindByPhone(ctx *gin.Context, u domain.User) (domain.User, error) {
	user, err := r.dao.FindByPhone(ctx, u.Phone)
	if err != nil {
		return domain.User{}, err
	}

	return r.entityToDomain(user), nil
}

func (r *CachedUserRepository) UpdateProfile(ctx *gin.Context, u domain.User) error {
	return r.dao.UpdateProfile(ctx, r.domainToEntity(u))
}

func (r *CachedUserRepository) UpdatePassword(ctx *gin.Context, u domain.User) error {
	return r.dao.UpdatePassword(ctx, r.domainToEntity(u))
}

func (r *CachedUserRepository) FindById(ctx *gin.Context, id int64) (domain.User, error) {
	// 先从Cache里找
	u, err := r.cache.Get(ctx, id)
	if err == nil {
		// 必然是有数据
		return u, nil
	}
	// 没这个数据
	//if err == cache.ErrKeyNotExist {
	// 去数据库里面加载
	//}

	user, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	u = r.entityToDomain(user)

	// 找到了， 放到cache里，考虑用go routine放
	go func() {
		r.cache.Set(ctx, u)
		if err != nil {
			// 我这里怎么办？
			// 打日志，做监控
			//return domain.User{}, err
		}
	}()

	return u, err
}

func (r *CachedUserRepository) FindByWechat(ctx *gin.Context, openId string) (domain.User, error) {
	user, err := r.dao.FindByWechat(ctx, openId)
	if err != nil {
		return domain.User{}, err
	}

	return r.entityToDomain(user), nil
}

func (r *CachedUserRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id:       u.Id,
		Email:    u.Email,
		Phone:    u.Phone,
		Password: u.Password,
		WechatOpenID: sql.NullString{
			String: u.WechatInfo.OpenID,
			Valid:  u.WechatInfo.OpenID != "",
		},
		WechatUnionID: sql.NullString{
			String: u.WechatInfo.UnionID,
			Valid:  u.WechatInfo.UnionID != "",
		},
		Ctime: u.Ctime.UnixMilli(),
	}
}

func (r *CachedUserRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		Phone:    u.Phone,
		WechatInfo: domain.WechatInfo{
			UnionID: u.WechatUnionID.String,
			OpenID:  u.WechatOpenID.String,
		},
		Ctime: time.UnixMilli(u.Ctime),
	}
}
