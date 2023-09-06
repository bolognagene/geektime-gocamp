package web

import (
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/service"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

/*
* User模块业务flow:
* /users/singup: 注册。 输入电话号码，邮箱，密码，确认密码。其中电话号码作为主账号
* /users/login:  登陆。 通过电话号码和密码登陆
* /users/profile:  显示账号主页信息。 显示昵称、生日和简介三个项目
* /users/profile/edit:  修改账号主页信息。修改昵称、生日和简介。
* /users/login/sms/send:  登陆业务发送短信。发送电话号码
* /users/login/sms/code:  登陆业务验证短信。发送电话号码和输入的验证码
* /users/edit_email:   修改邮箱。发送新邮箱地址
 */

const bizLogin = "login"

//const bizEditEmail = "editEmail"

// UserHandler 我准备在它上面定义跟用户有关的路由
type UserHandler struct {
	svc         *service.UserService
	smsSvc      *service.SMSCodeService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
	nickNameExp *regexp.Regexp
	birthdayExp *regexp.Regexp
	introExp    *regexp.Regexp
	phoneExp    *regexp.Regexp
}

func NewUserHandler(svc *service.UserService, smsSvc *service.SMSCodeService) *UserHandler {
	const (
		emailRegexPattern    = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
		passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
		// 昵称长度至多8个中文字符或者16个英文字符
		nickNameRegexPattern = `^[\s\S\u4E00-\u9FA5]{1,16}$`
		// 生日需要符合 YYYY-MM-DD 的格式
		birthdayRegexPattern = `^(19|20)\d{2}-(1[0-2]|0?[1-9])-(0?[1-9]|[1-2][0-9]|3[0-1])$`
		// 简介长度至多64个中文字符或者128个英文字符
		introRegexPattern = `^[\s\S\u4E00-\u9FA5]{1,128}$`
		phoneRegexPattern = `^1[3456789]\d{9}$`
	)
	emailExp := regexp.MustCompile(emailRegexPattern, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexPattern, regexp.None)
	nickNameExp := regexp.MustCompile(nickNameRegexPattern, regexp.None)
	birthdayExp := regexp.MustCompile(birthdayRegexPattern, regexp.None)
	introExp := regexp.MustCompile(introRegexPattern, regexp.None)
	phoneExp := regexp.MustCompile(phoneRegexPattern, regexp.None)
	return &UserHandler{
		svc:         svc,
		smsSvc:      smsSvc,
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
	ug.POST("/signup", u.SignUp)
	ug.POST("/login", u.LoginJWT)
	ug.GET("/profile", u.ProfileJWT)
	ug.POST("/profile/edit", u.EditJWT)
	ug.POST("/login/sms/send", u.SendLoginSMSCode)
	ug.POST("/login/sms/code", u.LoginSMS)
}

func (u *UserHandler) SignUp(ctx *gin.Context) {
	type SignUpReq struct {
		Phone           string `json:"phone"`
		Email           string `json:"email"`
		ConfirmPassword string `json:"confirmPassword"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	// Bind 方法会根据 Content-Type 来解析你的数据到 req 里面
	// 解析错了，就会直接写回一个 400 的错误
	if err := ctx.Bind(&req); err != nil {
		return
	}

	ok, err := u.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "你的电话格式不对",
		})
		return
	}

	ok, err = u.emailExp.MatchString(req.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "你的邮箱格式不对",
		})
		return
	}

	if req.ConfirmPassword != req.Password {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 4,
			Msg:  "两次输入的密码不一致",
		})
		return
	}
	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		// 记录日志
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "密码必须大于8位，包含英文字母、数字、特殊字符",
		})
		return
	}

	// 调用一下 svc 的方法
	err = u.svc.SignUp(ctx, domain.User{
		Phone:    req.Phone,
		Email:    req.Email,
		Password: req.Password,
	})
	if err == service.ErrUserDuplicateEmail {
		ctx.JSON(http.StatusConflict, Result{
			Code: 4,
			Msg:  "邮箱冲突",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "注册成功",
	})
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Phone, req.Password)
	if err == ErrInvalidUserOrPassword {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 4,
			Msg:  "用户名或密码不对",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	// 步骤2
	// 在这里登录成功了
	// 设置 session
	sess := sessions.Default(ctx)
	// 我可以随便设置值了
	// 你要放在 session 里面的值
	sess.Set("userId", user.Id)
	sess.Save()
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "登录成功",
	})
	return
}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	user, err := u.svc.Login(ctx, req.Phone, req.Password)
	if err == ErrInvalidUserOrPassword {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 4,
			Msg:  "用户名或密码不对",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	// 步骤2
	// 在这里用 JWT 设置登录态
	// 生成一个 JWT token
	err = u.SetJWTToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	fmt.Println(user)
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "登录成功",
	})
	return
}

func (u *UserHandler) SetJWTToken(ctx *gin.Context, uid int64) error {
	claims := service.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(), // Context里自带了UserAgent, 因此不用自己去得到
	}

	// 生成一个token
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	// 用这个32位字符进行签名，签名后的字串将放到resp里，因此是可以被用户看到的字串
	tokenStr, err := token.SignedString([]byte("hLU$fxHWHXp%ZiIIk8zG1mndXpE#n3EO"))
	if err != nil {
		return err
	}
	// 将这个tokenStr放到resp的Header里````
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (u *UserHandler) EditJWT(ctx *gin.Context) {
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
	ok, err := u.nickNameExp.MatchString(req.NickName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "昵称格式不对",
		})
		return
	}

	// 校验生日
	ok, err = u.birthdayExp.MatchString(req.Birthday)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "生日格式不对",
		})
		return
	}

	// 校验简介
	ok, err = u.introExp.MatchString(req.Intro)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "简介格式不对",
		})
		return
	}

	err = u.svc.EditJWT(ctx, domain.User{
		NickName: req.NickName,
		Birthday: req.Birthday,
		Intro:    req.Intro,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "更新用户信息成功",
	})
	return
}

func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	// JWT方式，从ctx里拿到claims
	up, err := u.svc.ProfileJWT(ctx)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  fmt.Sprintln("这是你的 Profile: \n 昵称是: " + up.NickName + ", 生日是:" + up.Birthday + ", 简介是: " + up.Intro),
	})
}

func (u *UserHandler) SendLoginSMSCode(ctx *gin.Context) {
	type SendSMSReq struct {
		Phone string `json:"phone"`
	}

	var req SendSMSReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 是不是一个合法的手机号码
	// 考虑正则表达式
	ok, err := u.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "电话号码格式不对",
		})
		return
	}

	err = u.smsSvc.Send(ctx, bizLogin, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Code: 0,
			Msg:  "发送成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusInternalServerError, Result{
			Msg: "发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
}

func (u *UserHandler) LoginSMS(ctx *gin.Context) {
	type LoginSMSReq struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	var req LoginSMSReq
	if err := ctx.Bind(&req); err != nil {
		return
	}

	// 这里是Phone的校验
	ok, err := u.phoneExp.MatchString(req.Phone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusBadRequest, Result{
			Code: 4,
			Msg:  "电话号码格式不对",
		})
		return
	}

	// 这里是code的校验
	codeRegexPattern := `^\d{6}$`
	codeExp := regexp.MustCompile(codeRegexPattern, regexp.None)
	ok, err = codeExp.MatchString(req.Code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 4,
			Msg:  "验证码有误",
		})
		return
	}

	// 开始校验
	ok, err = u.smsSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusUnauthorized, Result{
			Code: 4,
			Msg:  "验证码有误",
		})
		return
	}

	// 校验通过，设置登陆态
	// 由于登陆态需要知道user.id，所以这里我们还得查一下user.id
	du, err := u.svc.FindByPhone(ctx, req.Phone)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}

	// 设置登陆态
	err = u.SetJWTToken(ctx, du.Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	fmt.Println(du)
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "登录成功",
	})
	return
}
