package domain

import "time"

// UserProfile 领域对象，保存user的一些次要信息
type UserProfile struct {
	UserId   int64 //与User表里的ID对应
	NickName string
	Birthday string
	Intro    string
	Ctime    time.Time
}
