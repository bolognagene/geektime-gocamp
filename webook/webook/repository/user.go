package repository

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/dao"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository/dao/cache"
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.UserCache
}

func NewUserRepository(dao *dao.UserDAO, c *cache.UserCache) *UserRepository {
	return &UserRepository{
		dao:   dao,
		cache: c,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	// SELECT * FROM `users` WHERE `email`=?
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) FindById(ctx context.Context, uid int64) (domain.User, error) {
	//先从cache里找
	du, err := r.cache.Get(ctx, uid)
	if err != nil {
		// 必然是有数据
		return du, nil
	}

	// SELECT * FROM `users` WHERE `id`=?
	u, err := r.dao.FindById(ctx, uid)
	if err != nil {
		return domain.User{}, err
	}
	du = domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		Birthday: u.Birthday,
		NickName: u.NickName,
		Intro:    u.Intro,
	}

	// 写入cache
	go func() {
		r.cache.Set(ctx, du)
		if err != nil {
			// 我这里怎么办？
			// 打日志，做监控
			//return domain.User{}, err
		}
	}()

	return du, err
}
func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	err := r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
		NickName: u.NickName,
		Birthday: u.Birthday,
		Intro:    u.Intro,
	})
	if err != nil {
		return err
	}

	// 写入cache
	go func() {
		r.cache.Set(ctx, u)
	}()

	return err
}

func (r *UserRepository) UpdateById(ctx context.Context, u domain.User) error {
	err := r.dao.UpdateById(ctx, dao.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
		NickName: u.NickName,
		Birthday: u.Birthday,
		Intro:    u.Intro,
	})
	if err != nil {
		return err
	}

	// 写入cache
	go func() {
		r.cache.Set(ctx, u)
	}()

	return err
}
