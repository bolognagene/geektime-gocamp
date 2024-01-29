package repository

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/cache"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository/dao"
	"github.com/gin-gonic/gin"
)

var (
	ErrUserDuplicatePhone = dao.ErrUserDuplicatePhone
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *UserRepository) Create(ctx *gin.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Phone:    u.Phone,
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *UserRepository) FindByEmail(ctx *gin.Context, u domain.User) (domain.User, error) {
	user, err := r.dao.FindByEmail(ctx, u.Email)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:       user.Id,
		Email:    user.Email,
		Password: user.Password,
	}, nil
}

func (r *UserRepository) FindByPhone(ctx *gin.Context, u domain.User) (domain.User, error) {
	user, err := r.dao.FindByPhone(ctx, u.Phone)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Id:       user.Id,
		Phone:    user.Phone,
		Email:    user.Email,
		Password: user.Password,
	}, nil
}

func (r *UserRepository) UpdateProfile(ctx *gin.Context, u domain.User) error {
	return r.dao.UpdateProfile(ctx, dao.User{
		Id:       u.Id,
		Nickname: u.Nickname,
		Birthday: u.Birthday,
		Intro:    u.Intro,
	})
}

func (r *UserRepository) UpdatePassword(ctx *gin.Context, u domain.User) error {
	return r.dao.UpdatePassword(ctx, dao.User{
		Id:       u.Id,
		Password: u.Password,
	})
}

func (r *UserRepository) FindById(ctx *gin.Context, id int64) (domain.User, error) {
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

	u = domain.User{
		Id:       user.Id,
		Password: user.Password,
		Nickname: user.Nickname,
		Birthday: user.Birthday,
		Intro:    user.Intro,
		Email:    user.Email,
	}

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
