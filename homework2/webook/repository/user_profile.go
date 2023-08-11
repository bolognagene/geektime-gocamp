package repository

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/repository/dao"
)

type UserProfileRepository struct {
	dao *dao.UserProfileDao
}

func NewUserProfileRepository(dao *dao.UserProfileDao) *UserProfileRepository {
	return &UserProfileRepository{
		dao: dao,
	}
}

func (r *UserProfileRepository) Create(ctx context.Context, up domain.UserProfile) error {
	return r.dao.Insert(ctx, dao.UserProfile{
		UserId:   up.UserId,
		NickName: up.NickName,
		Birthday: up.Birthday,
		Intro:    up.Intro,
	})
}

func (r *UserProfileRepository) FindByUserId(ctx context.Context, uid int64) (domain.UserProfile, error) {
	// SELECT * FROM `user_profile` WHERE `userid`=?
	up, err := r.dao.FindByUserId(ctx, uid)
	if err != nil {
		return domain.UserProfile{}, err
	}
	return domain.UserProfile{
		UserId:   up.UserId,
		NickName: up.NickName,
		Birthday: up.Birthday,
		Intro:    up.Intro,
	}, nil
}

func (r *UserProfileRepository) Update(ctx context.Context, up domain.UserProfile) error {
	/*return r.dao.Update(ctx, dao.UserProfile{
		UserId:   up.UserId,
		NickName: up.NickName,
		Birthday: up.Birthday,
		Intro:    up.Intro,
	})*/

	return r.dao.Update(ctx, up)
}
