package service

import (
	"errors"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicatePhone    = repository.ErrUserDuplicatePhone
	ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) SignUp(ctx *gin.Context, u domain.User) error {
	// 你要考虑加密放在哪里的问题了
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	// 然后就是，存起来
	return s.repo.Create(ctx, u)
}

func (s *UserService) FindOrCreate(ctx *gin.Context, u domain.User) (domain.User, error) {
	// 这个叫做快路径
	user, err := s.repo.FindByPhone(ctx, u)
	if err != repository.ErrUserNotFound && user.Id != 0 {
		// 绝大部分请求进来这里
		// nil 会进来这里
		// 不为 ErrUserNotFound 的也会进来这里
		return user, err
	}

	// 在系统资源不足，触发降级之后，不执行慢路径了
	//if ctx.Value("降级") == "true" {
	//	return domain.User{}, errors.New("系统降级了")
	//}

	// 没有找到，需要创建这个用户
	// 这个叫做慢路径
	err = s.repo.Create(ctx, u)
	if err != nil {
		return u, err
	}

	// 要得到userId, 所以要重新找一遍
	// 因为这里会遇到主从延迟的问题
	return s.repo.FindByPhone(ctx, u)

}

func (s *UserService) Login(ctx *gin.Context, u domain.User) (domain.User, error) {
	user, err := s.repo.FindByPhone(ctx, u)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}

	// 比较密码了
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))
	if err != nil {
		// 这里可以加入DEBUG信息
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return user, nil
}

func (s *UserService) EditProfile(ctx *gin.Context, u domain.User) error {

	return s.repo.UpdateProfile(ctx, u)
}

func (s *UserService) EditPassword(ctx *gin.Context, u domain.User) error {
	// 你要考虑加密放在哪里的问题了
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)

	return s.repo.UpdatePassword(ctx, u)
}

func (s *UserService) Profile(ctx *gin.Context, id int64) (domain.User, error) {

	user, err := s.repo.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}
