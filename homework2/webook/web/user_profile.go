package web

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/homework2/webook/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserProfileHandler struct {
	svc         *service.UserProfileService
	nickNameExp *regexp.Regexp
	birthdayExp *regexp.Regexp
	introExp    *regexp.Regexp
}

func NewUserProfileHandler(svc *service.UserProfileService) *UserProfileHandler {
	const (
		// 昵称长度至多8个中文字符或者16个英文字符
		nickNameRegexPattern = `^[\u4E00-\u9FA5]{1,8}$|^[A-Za-z]{1,16}$`
		// 生日需要符合 YYYY-MM-DD 的格式
		birthdayRegexPattern = `^(19|20)\d{2}-(1[0-2]|0?[1-9])-(0?[1-9]|[1-2][0-9]|3[0-1])$`
		// 简介长度至多64个中文字符或者128个英文字符
		introRegexPattern = `^[\u4E00-\u9FA5]{1,64}$|^[A-Za-z]{1,128}$`
	)
	nickNameExp := regexp.MustCompile(nickNameRegexPattern, regexp.None)
	birthdayExp := regexp.MustCompile(birthdayRegexPattern, regexp.None)
	introExp := regexp.MustCompile(introRegexPattern, regexp.None)
	return &UserProfileHandler{
		svc:         svc,
		nickNameExp: nickNameExp,
		birthdayExp: birthdayExp,
		introExp:    introExp,
	}
}

func (uph *UserProfileHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users/profile")
	ug.GET("/", uph.Profile)
	ug.POST("/edit", uph.Edit)
}

func (uph *UserProfileHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		NickName string `json:"nick_name"`
		Birthday string `json:"birthday"`
		Intro    string `json:"intro"`
	}

	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 校验昵称
	ok, err := uph.nickNameExp.MatchString(req.NickName)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "昵称格式不对")
		return
	}

	// 校验生日
	ok, err = uph.birthdayExp.MatchString(req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "生日格式不对")
		return
	}

	// 校验简介
	ok, err = uph.introExp.MatchString(req.Intro)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "简介格式不对")
		return
	}

	err = uph.svc.Edit(ctx, domain.UserProfile{
		NickName: req.NickName,
		Birthday: req.Birthday,
		Intro:    req.Intro,
	})

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "更新用户信息成功")
	return
}

func (uph *UserProfileHandler) Profile(ctx *gin.Context) {
	// 从session里拿到userId，然后从数据库里查出来后直接打印出来
	up, err := uph.svc.Profile(ctx)
	if err != nil {
		return
	}

	ctx.String(http.StatusOK, "这是你的 Profile: \n 昵称是: "+up.NickName+", 生日是:"+up.Birthday+", 简介是"+up.Intro)
}
