package service

import (
	"context"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
	// 先找用户
	u, err := svc.repo.FindByEmail(ctx, email)
	if err == ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 比较密码了
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		// DEBUG
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	// 你要考虑加密放在哪里的问题了
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	// 然后就是，存起来
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) EditJWT(ctx *gin.Context, u domain.User) error {
	uid, err := getUserIdByJWT(ctx)
	if err != nil || uid == nil {
		return err
	}

	u.Id = uid.(int64)
	err = svc.repo.UpdateById(ctx, u)
	if err != nil {
		return err
	}

	return nil

}

func (svc *UserService) ProfileJWT(ctx *gin.Context) (domain.User, error) {
	uid, err := getUserIdByJWT(ctx)
	if err != nil || uid == nil {
		return domain.User{}, err
	}

	u, err := svc.repo.FindById(ctx, uid.(int64))

	if err != nil {
		return domain.User{}, err
	}

	return u, nil

}

// 用JWT的方式来取user id
func getUserIdByJWT(ctx *gin.Context) (any, error) {
	// 需要从Context里拿到Claims，然后再拿到user id
	c, _ := ctx.Get("claims")
	claims, ok := c.(*UserClaims)
	// ok 代表是不是 *UserClaims
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return nil, ErrNotLogin
	}
	println(claims.Uid)
	// claims取到了，然后取user id
	return claims.Uid, nil
}

type UserClaims struct {
	jwt.RegisteredClaims
	// 声明你自己的要放进去 token 里面的数据，
	// user id
	Uid int64
	// user agent
	UserAgent string
}
