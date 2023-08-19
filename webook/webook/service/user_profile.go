package service

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/repository"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

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

	up.UserId = uid.(int64)
	err = svc.repo.Update(ctx, up)
	if err != nil {
		return err
	}

	return nil

}

func (svc *UserProfileService) Profile(ctx *gin.Context) (domain.UserProfile, error) {
	var upnil domain.UserProfile
	uid, err := getUserId(ctx)
	if err != nil {
		return upnil, err
	}

	up, err := svc.repo.FindByUserId(ctx, uid.(int64))
	if err == ErrUserNotFound {
		// User profile没有找到是正常现象，因此直接返回nil的错误
		return upnil, nil
	}
	if err != nil {
		return upnil, err
	}

	return up, nil

}
