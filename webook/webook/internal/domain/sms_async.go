package domain

import "time"

type SMSAsync struct {
	Id      int64
	Biz     string
	Args    string
	Numbers string
	Ctime   time.Time
}
