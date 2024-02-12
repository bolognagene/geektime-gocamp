package web

import (
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	myjwt "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web/jwt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

const biz = "login"

type UserHandler struct {
	svc         service.UserService
	codeSvc     service.CodeService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	nickNameExp *regexp.Regexp
	birthdayExp *regexp.Regexp
	introExp    *regexp.Regexp
	phoneExp    *regexp.Regexp
	myjwt.JwtHandler
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
		// 昵称长度至多8个中文字符或者16个英文字符
		nickNameRegexPattern = `^[\s\S\u4E00-\u9FA5]{1,16}$`
		// 生日需要符合 YYYY-MM-DD 的格式
		birthdayRegexPattern = `^(19|20)\d{2}-(1[0-2]|0?[1-9])-(0?[1-9]|[1-2][0-9]|3[0-1])$`
		// 简介长度至多64个中文字符或者128个英文字符
		introRegexPattern = `^[\s\S\u4E00-\u9FA5]{1,128}$`
		phoneRegexPattern = `^1[3-9]\d{9}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	nickNameExp := regexp.MustCompile(nickNameRegexPattern, regexp.None)
	birthdayExp := regexp.MustCompile(birthdayRegexPattern, regexp.None)
	introExp := regexp.MustCompile(introRegexPattern, regexp.None)
	phoneExp := regexp.MustCompile(phoneRegexPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		codeSvc:     codeSvc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
		nickNameExp: nickNameExp,
		birthdayExp: birthdayExp,
		introExp:    introExp,
		phoneExp:    phoneExp,
	}
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	//ug.GET("/profile", u.Profile)
	ug.GET("/profile", u.ProfileByJWT)
	ug.POST("/signup", u.SignUp)
	//ug.POST("/login", u.Login)
	ug.POST("/login", u.LoginByJWT)
	ug.POST("/logout", u.LogoutByJWT)
	//ug.POST("/profile/edit", u.EditProfile)
	ug.POST("/profile/edit", u.EditProfileByJWT)
	//ug.POST("/edit", u.EditPassword)
	ug.POST("/edit", u.EditPasswordByJWT)
	ug.POST("/login_sms/code/send", u.SendLoginSMSCode)
	ug.POST("/login_sms", u.LoginBySMS)
	ug.POST("/refresh_token", u.RefreshToken)
}

func (u *UserHandler) Profile(ctx *gin.Context) {

	sess := sessions.Default(ctx)
	id := sess.Get("userId")

	user, err := u.svc.Profile(ctx, id.(int64))
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "这是我的 Profile: \n 昵称是: "+user.Nickname+", 生日是:"+user.Birthday+", 简介是: "+user.Intro)
}

func (u *UserHandler) ProfileByJWT(ctx *gin.Context) {
	claims := u.GetUserClaim(ctx)
	if claims == nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	id := claims.Uid
	user, err := u.svc.Profile(ctx, id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "这是我的 Profile: \n 昵称是: "+user.Nickname+", 生日是:"+user.Birthday+", 简介是: "+user.Intro)
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Phone           string `json:"phone"`
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq

	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	ok, err := u.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "你的电话号码格式不对",
		})
		return
	}

	ok, err = u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "你的邮箱格式不对",
		})
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "两次输入的密码不一致",
		})
		return
	}

	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "密码必须大于8位，包含数字、特殊字符",
		})
		return
	}

	//ctx.String(http.StatusOK, "注册成功")
	// %v	the value in a default format
	//	when printing structs, the plus flag (%+v) adds field names
	//fmt.Printf("%v", req)
	// 这边就是数据库操作
	err = u.svc.SignUp(ctx, domain.User{
		Phone:    req.Phone,
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrUserDuplicate {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "电话号码冲突",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "注册成功",
	})
	return
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginReq

	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	user, err := u.svc.Login(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码不对")
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 步骤2
	// 在这里登录成功了
	// 设置 session
	sess := sessions.Default(ctx)
	// 我可以随便设置值了
	// 你要放在 session 里面的值
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		//Secure:   true, //只在生产环境上设置true
		//HttpOnly: true, //只在生产环境上设置true
		// 一分钟过期
		MaxAge: 60,
	})
	sess.Save()
	ctx.String(http.StatusOK, "登录成功")
	return

}

func (u *UserHandler) LoginByJWT(ctx *gin.Context) {
	type LoginReq struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}

	var req LoginReq

	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	user, err := u.svc.Login(ctx, domain.User{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err == service.ErrInvalidUserOrPassword {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "用户名或密码不对",
		})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	err = u.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.String(http.StatusOK, "登录成功")
	return
}

func (u *UserHandler) LogoutByJWT(ctx *gin.Context) {
	err := u.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "登出失败",
		})
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "登出成功",
	})
}

func (u *UserHandler) LoginBySMS(ctx *gin.Context) {
	type LoginReq struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	var req LoginReq

	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	ok, err := u.codeSvc.Verify(ctx, biz, req.Phone, req.Code)
	if err == service.ErrCodeVerifyTooManyTimes {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码输错过多",
		})
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "验证码错误",
		})
	}

	// 我这个手机号，会不会是一个新用户呢？
	// 这样子
	user, err := u.svc.FindOrCreate(ctx, domain.User{
		Phone: req.Phone,
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	err = u.SetLoginToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.String(http.StatusOK, "登录成功")
	return
}

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type LoginReq struct {
		Phone string `json:"phone"`
	}

	var req LoginReq

	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	err = u.codeSvc.Send(ctx, biz, req.Phone)
	if err != nil {
		if err == service.ErrCodeSendTooMany {
			ctx.JSON(http.StatusOK, Result{
				Code: 4,
				Msg:  "验证码发送太频繁",
			})
		} else {
			ctx.JSON(http.StatusOK, Result{
				Code: 5,
				Msg:  "系统错误",
			})
		}
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 2,
		Msg:  "验证码发送成功",
	})
}

func (u *UserHandler) EditProfile(ctx *gin.Context) {
	type EditProfileReq struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		Intro    string `json:"intro"`
	}

	var req EditProfileReq

	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	// 校验昵称
	ok, err := u.nickNameExp.MatchString(req.Nickname)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "昵称格式不对")
		return
	}

	// 校验生日
	ok, err = u.birthdayExp.MatchString(req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "生日格式不对")
		return
	}

	// 校验简介
	ok, err = u.introExp.MatchString(req.Intro)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "简介格式不对")
		return
	}

	// Session方式
	sess := sessions.Default(ctx)
	id := sess.Get("userId")

	err = u.svc.EditProfile(ctx, domain.User{
		Id:       id.(int64),
		Birthday: req.Birthday,
		Nickname: req.Nickname,
		Intro:    req.Intro,
	})

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "更新用户信息成功")
	return
}

func (u *UserHandler) EditProfileByJWT(ctx *gin.Context) {
	type EditProfileReq struct {
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		Intro    string `json:"intro"`
	}

	var req EditProfileReq

	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	// 校验昵称
	ok, err := u.nickNameExp.MatchString(req.Nickname)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "昵称格式不对")
		return
	}

	// 校验生日
	ok, err = u.birthdayExp.MatchString(req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "生日格式不对")
		return
	}

	// 校验简介
	ok, err = u.introExp.MatchString(req.Intro)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "简介格式不对")
		return
	}

	claims := u.GetUserClaim(ctx)
	if claims == nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	id := claims.Uid
	err = u.svc.EditProfile(ctx, domain.User{
		Id:       id,
		Birthday: req.Birthday,
		Nickname: req.Nickname,
		Intro:    req.Intro,
	})

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "更新用户信息成功")
	return
}

func (u *UserHandler) EditPassword(ctx *gin.Context) {
	type EditPasswordReq struct {
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req EditPasswordReq

	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入的密码不一致")
		return
	}

	ok, err := u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码必须大于8位，包含数字、特殊字符")
		return
	}

	sess := sessions.Default(ctx)
	id := sess.Get("userId")

	err = u.svc.EditPassword(ctx, domain.User{
		Id:       id.(int64),
		Password: req.Password,
	})
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.String(http.StatusOK, "更改密码成功")
	return
}

func (u *UserHandler) EditPasswordByJWT(ctx *gin.Context) {
	type EditPasswordReq struct {
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req EditPasswordReq

	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次输入的密码不一致")
		return
	}

	ok, err := u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码必须大于8位，包含数字、特殊字符")
		return
	}

	claims := u.GetUserClaim(ctx)
	if claims == nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	id := claims.Uid
	err = u.svc.EditPassword(ctx, domain.User{
		Id:       id,
		Password: req.Password,
	})
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.String(http.StatusOK, "更改密码成功")
	return
}

func (u *UserHandler) RefreshToken(ctx *gin.Context) {
	refreshToken := u.ExtractToken(ctx)
	var rc myjwt.RefreshClaims
	// 先验证refreshToken
	token, err := jwt.ParseWithClaims(refreshToken, &rc, func(token *jwt.Token) (interface{}, error) {
		return u.GetAtKey(ctx), nil
	})
	if err != nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}

	// 在验证是否登出，看ssid是否在redis里
	if u.CheckSession(ctx, rc.Ssid) {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}

	// 搞个新的access_token
	err = u.SetJWTToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "刷新成功",
	})
}
