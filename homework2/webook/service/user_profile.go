package service

import (
	"errors"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/repository"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

var ErrNotLogin = errors.New("还没有登录")

type UserProfileService struct {
	repo *repository.UserProfileRepository
}

func NewUserProfileService(repo *repository.UserProfileRepository) *UserProfileService {
	return &UserProfileService{
		repo: repo,
	}
}
func getUserId(ctx *gin.Context) (any, error) {
	// 需要从session里拿到userId
	sess := sessions.Default(ctx)
	uid := sess.Get("userId")
	if uid == nil {
		// 没有登录
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return nil, ErrNotLogin
	}
	return uid, nil
}

func (svc *UserProfileService) Edit(ctx *gin.Context, up domain.UserProfile) error {
	uid, err := getUserId(ctx)
	if err != nil {
		return err
	}

	// 根据userId来查询user_profile表，如果已经有这条记录就调用Update，否则调用Create
	upp, err := svc.repo.FindById(ctx, uid.(int64))
	// 如果没找到，就创建这条记录
	if err == repository.ErrUserNotFound {
		createErr := svc.repo.Create(ctx, up)
		if createErr != nil {
			return createErr
		}

		return nil
	}

	// 其他错误直接返回
	if err != nil {
		return err
	}

	// 如果成功找到UserProfile, 则更新改记录
	updateError := svc.repo.Update(ctx, domain.UserProfile{
		UserId:   upp.UserId,
		NickName: up.NickName,
		Birthday: up.Birthday,
		Intro:    up.Intro,
	})

	if updateError != nil {
		return updateError
	}

	return nil

}

func (svc *UserProfileService) Profile(ctx *gin.Context) (domain.UserProfile, error) {
	var upnil domain.UserProfile
	uid, err := getUserId(ctx)
	if err != nil {
		return upnil, err
	}

	up, err := svc.repo.FindById(ctx, uid.(int64))
	if err != nil {
		return up, err
	}

	return up, nil

}
