package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/domain"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service"
	svcmocks "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/service/mocks"
	ijwt "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web/jwt"
	jwtmocks "github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/web/jwt/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEncrypt(t *testing.T) {
	_ = NewUserHandler(nil, nil, nil)
	password := "hello#world123"
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	err = bcrypt.CompareHashAndPassword(encrypted, []byte(password))
	assert.NoError(t, err)
}

func TestNil(t *testing.T) {
	testTypeAssert(nil)
}

func testTypeAssert(c any) {
	_, ok := c.(*ijwt.UserClaims)
	println(ok)
}

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) service.UserService

		reqBody string

		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(nil)
				// 注册成功是 return nil
				return usersvc
			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123",
	"confirmPassword": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			name: "参数不对，bind 失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				// 注册成功是 return nil
				return usersvc
			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123"
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},

			reqBody: `
{
	"email": "123@q",
	"password": "hello#world123",
	"confirmPassword": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "你的邮箱格式不对",
		},
		{
			name: "两次输入密码不匹配",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world1234",
	"confirmPassword": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "两次输入的密码不一致",
		},
		{
			name: "密码格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				return usersvc
			},
			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello123",
	"confirmPassword": "hello123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "密码必须大于8位，包含数字、特殊字符",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(service.ErrUserDuplicateEmail)
				// 注册成功是 return nil
				return usersvc
			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123",
	"confirmPassword": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "邮箱冲突",
		},
		{
			name: "系统异常",
			mock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(errors.New("随便一个 error"))
				// 注册成功是 return nil
				return usersvc
			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123",
	"confirmPassword": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "系统异常",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			// 用不上 codeSvc
			h := NewUserHandler(tc.mock(ctrl), nil, nil)
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost,
				"/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			// 这里你就可以继续使用 req

			resp := httptest.NewRecorder()
			// 这就是 HTTP 请求进去 GIN 框架的入口。
			// 当你这样调用的时候，GIN 就会处理这个请求
			// 响应写回到 resp 里
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())

		})
	}
}

func TestMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usersvc := svcmocks.NewMockUserService(ctrl)

	usersvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).
		Return(errors.New("mock error"))

	//usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
	//	Email: "124@qq.com",
	//}).Return(errors.New("mock error"))

	err := usersvc.SignUp(context.Background(), domain.User{
		Email: "123@qq.com",
	})
	t.Log(err)
}

func TestUserHandler_LoginJWT(t *testing.T) {
	testCases := []struct {
		name string

		usrSvcMock func(ctrl *gomock.Controller) service.UserService

		jwtHandlerMock func(ctrl *gomock.Controller) ijwt.Handler

		reqBody string

		wantCode int
		wantBody string
	}{
		{
			name: "登陆成功",
			usrSvcMock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().Login(gomock.Any(),
					"123@qq.com", "hello#world123").Return(domain.User{}, nil)
				// 登陆成功是 return nil
				return usersvc
			},

			jwtHandlerMock: func(ctrl *gomock.Controller) ijwt.Handler {
				jwtHander := jwtmocks.NewMockHandler(ctrl)
				var a int64 = 0
				jwtHander.EXPECT().SetLoginToken(gomock.Any(), a).Return(nil)

				return jwtHander
			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "登录成功",
		},
		{
			name: "参数不对，bind 失败",
			usrSvcMock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)

				return usersvc
			},

			jwtHandlerMock: func(ctrl *gomock.Controller) ijwt.Handler {
				jwtHander := jwtmocks.NewMockHandler(ctrl)

				return jwtHander
			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123"
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "设置token系统错误",
			usrSvcMock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().Login(gomock.Any(),
					"123@qq.com", "hello#world123").Return(domain.User{}, nil)

				return usersvc
			},

			jwtHandlerMock: func(ctrl *gomock.Controller) ijwt.Handler {
				jwtHander := jwtmocks.NewMockHandler(ctrl)
				var a int64 = 0
				jwtHander.EXPECT().SetLoginToken(gomock.Any(), a).Return(errors.New("随便一个 error"))

				return jwtHander
			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "系统错误",
		},
		{
			name: "登陆系统错误",
			usrSvcMock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().Login(gomock.Any(),
					"123@qq.com", "hello#world123").Return(domain.User{}, errors.New("随便一个 error"))

				return usersvc
			},

			jwtHandlerMock: func(ctrl *gomock.Controller) ijwt.Handler {
				jwtHander := jwtmocks.NewMockHandler(ctrl)

				return jwtHander
			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "系统错误",
		},
		{
			name: "登陆用户名或密码不对",
			usrSvcMock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().Login(gomock.Any(),
					"123@qq.com", "hello#world123").Return(domain.User{}, service.ErrInvalidUserOrPassword)

				return usersvc
			},

			jwtHandlerMock: func(ctrl *gomock.Controller) ijwt.Handler {
				jwtHander := jwtmocks.NewMockHandler(ctrl)

				return jwtHander
			},

			reqBody: `
{
	"email": "123@qq.com",
	"password": "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "用户名或密码不对",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			// 用不上 codeSvc
			h := NewUserHandler(tc.usrSvcMock(ctrl), nil, tc.jwtHandlerMock(ctrl))
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost,
				"/users/login", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			// 这里你就可以继续使用 req

			resp := httptest.NewRecorder()
			// 这就是 HTTP 请求进去 GIN 框架的入口。
			// 当你这样调用的时候，GIN 就会处理这个请求
			// 响应写回到 resp 里
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())

		})
	}
}

func TestUserHandler_LoginSMS(t *testing.T) {
	testCases := []struct {
		name string

		usrSvcMock func(ctrl *gomock.Controller) service.UserService

		codeSvcMock func(ctrl *gomock.Controller) service.CodeService

		jwtHandlerMock func(ctrl *gomock.Controller) ijwt.Handler

		reqBody string

		wantCode int
		wantBody Result
	}{
		{
			name: "验证码校验通过",
			usrSvcMock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().FindOrCreate(gomock.Any(),
					"15812345678").Return(domain.User{
					Id: 1,
				}, nil)

				return usersvc
			},

			codeSvcMock: func(ctrl *gomock.Controller) service.CodeService {
				codesvc := svcmocks.NewMockCodeService(ctrl)
				codesvc.EXPECT().Verify(gomock.Any(),
					"login", "15812345678",
					"123456").Return(true, nil)

				return codesvc
			},

			jwtHandlerMock: func(ctrl *gomock.Controller) ijwt.Handler {
				jwtHander := jwtmocks.NewMockHandler(ctrl)
				var a int64 = 1
				jwtHander.EXPECT().SetLoginToken(gomock.Any(), a).Return(nil)

				return jwtHander
			},

			reqBody: `
{
	"phone": "15812345678",
	"code": "123456"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Msg: "验证码校验通过",
			},
		},
		{
			name: "参数不对，bind 失败",
			usrSvcMock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)

				return usersvc
			},

			codeSvcMock: func(ctrl *gomock.Controller) service.CodeService {
				codesvc := svcmocks.NewMockCodeService(ctrl)

				return codesvc
			},

			jwtHandlerMock: func(ctrl *gomock.Controller) ijwt.Handler {
				jwtHander := jwtmocks.NewMockHandler(ctrl)

				return jwtHander
			},

			reqBody: `
{
	"phone": "15812345678",
	"code": "123456"
`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "设置token系统错误",
			usrSvcMock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().FindOrCreate(gomock.Any(),
					"15812345678").Return(domain.User{
					Id: 1,
				}, nil)

				return usersvc
			},

			codeSvcMock: func(ctrl *gomock.Controller) service.CodeService {
				codesvc := svcmocks.NewMockCodeService(ctrl)
				codesvc.EXPECT().Verify(gomock.Any(),
					"login", "15812345678",
					"123456").Return(true, nil)

				return codesvc
			},

			jwtHandlerMock: func(ctrl *gomock.Controller) ijwt.Handler {
				jwtHander := jwtmocks.NewMockHandler(ctrl)
				var a int64 = 1
				jwtHander.EXPECT().SetLoginToken(gomock.Any(), a).Return(errors.New("随便一个 error"))

				return jwtHander
			},

			reqBody: `
{
	"phone": "15812345678",
	"code": "123456"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "FindOrCreate系统错误",
			usrSvcMock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().FindOrCreate(gomock.Any(),
					"15812345678").Return(domain.User{}, errors.New("随便一个 error"))

				return usersvc
			},

			codeSvcMock: func(ctrl *gomock.Controller) service.CodeService {
				codesvc := svcmocks.NewMockCodeService(ctrl)
				codesvc.EXPECT().Verify(gomock.Any(),
					"login", "15812345678",
					"123456").Return(true, nil)

				return codesvc
			},

			jwtHandlerMock: func(ctrl *gomock.Controller) ijwt.Handler {
				jwtHander := jwtmocks.NewMockHandler(ctrl)

				return jwtHander
			},

			reqBody: `
{
	"phone": "15812345678",
	"code": "123456"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "验证码有误",
			usrSvcMock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)

				return usersvc
			},

			codeSvcMock: func(ctrl *gomock.Controller) service.CodeService {
				codesvc := svcmocks.NewMockCodeService(ctrl)
				codesvc.EXPECT().Verify(gomock.Any(),
					"login", "15812345678",
					"123456").Return(false, nil)

				return codesvc
			},

			jwtHandlerMock: func(ctrl *gomock.Controller) ijwt.Handler {
				jwtHander := jwtmocks.NewMockHandler(ctrl)

				return jwtHander
			},

			reqBody: `
{
	"phone": "15812345678",
	"code": "123456"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 4,
				Msg:  "验证码有误",
			},
		},
		{
			name: "Verify系统错误",
			usrSvcMock: func(ctrl *gomock.Controller) service.UserService {
				usersvc := svcmocks.NewMockUserService(ctrl)

				return usersvc
			},

			codeSvcMock: func(ctrl *gomock.Controller) service.CodeService {
				codesvc := svcmocks.NewMockCodeService(ctrl)
				codesvc.EXPECT().Verify(gomock.Any(),
					"login", "15812345678",
					"123456").Return(false, errors.New("随便一个 error"))

				return codesvc
			},

			jwtHandlerMock: func(ctrl *gomock.Controller) ijwt.Handler {
				jwtHander := jwtmocks.NewMockHandler(ctrl)

				return jwtHander
			},

			reqBody: `
{
	"phone": "15812345678",
	"code": "123456"
}
`,
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			server := gin.Default()
			// 用不上 codeSvc
			h := NewUserHandler(tc.usrSvcMock(ctrl), tc.codeSvcMock(ctrl), tc.jwtHandlerMock(ctrl))
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost,
				"/users/login_sms", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			// 这里你就可以继续使用 req

			resp := httptest.NewRecorder()
			// 这就是 HTTP 请求进去 GIN 框架的入口。
			// 当你这样调用的时候，GIN 就会处理这个请求
			// 响应写回到 resp 里
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			var respJson Result
			json.Unmarshal([]byte(resp.Body.String()), &respJson)
			assert.Equal(t, tc.wantBody, respJson)

		})
	}
}
