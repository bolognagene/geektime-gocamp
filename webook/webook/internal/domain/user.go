package domain

import "time"

// User 领域对象，是 DDD 中的 entity
// BO(business object)
type User struct {
	Id       int64
	Phone    string
	Email    string
	Password string
	Nickname string
	Birthday string
	Intro    string
	Ctime    time.Time
}
