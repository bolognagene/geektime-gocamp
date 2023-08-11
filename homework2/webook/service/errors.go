package service

import (
	"errors"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/repository"
)

// 这种error方式很啰嗦，得研究下真实项目里error如何组织
var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrUserNotFound          = repository.ErrUserNotFound
	ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")
	ErrNotLogin              = errors.New("还没有登录")
)
