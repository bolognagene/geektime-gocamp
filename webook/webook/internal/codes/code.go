package codes

const (
	// CommonInvalidInput 任何模块都可以使用的表达输入错误
	CommonInvalidInput   = 400001
	CommonInternalServer = 500001
)

// 用户模块，模块码 01
const (
	UserOK = 201001
	// UserInvalidInput 用户模块输入错误，这是一个含糊的错误
	UserInvalidInput = 401001
	// UserInvalidOrPassword 用户不存在或者密码错误，这个你要小心，
	// 防止有人跟你过不去
	UserInvalidOrPassword = 401002
	// 验证码输错过多
	UserTooManyVerifiedFailed = 401003
	// 验证码发送太频繁
	UserTooManySendSMS = 401004
	// 无权限
	UserUnauthorized = 401005
	// 系统错误
	UserInternalServerError = 501001
)

// 文章模块， 模块代码02
const (
	ArticleOK                  = 202001
	ArticleInvalidInput        = 402001
	ArticleInternalServerError = 502001
)

var (
	// UserInvalidInputV1 这个东西是你 DEBUG 用的，不是给 C 端用户用的
	UserInvalidInputV1 = Code{
		Number: 401001,
		Msg:    "用户输入错误",
	}
)

type Code struct {
	Number int
	Msg    string
}
