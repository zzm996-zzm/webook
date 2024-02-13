package web

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/internal/domain"
	"webook/internal/service"
	svcmocks "webook/internal/service/mock"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name string
		// mock
		mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService)
		//构造请求，我需要的请求
		req func(t *testing.T) *http.Request

		//预期返回
		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(nil)
				return userSvc, nil
			},
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
"email": "123@qq.com",
"password": "hello#world123",
"confirmPassword": "hello#world123"
}`)))

				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: 200,
			wantBody: "注册成功",
		},
		{
			name: "Bind出错",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
"email": "123@qq.com",
"password": "hello#world
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},

			wantCode: http.StatusBadRequest,
			wantBody: "系统错误",
		},
		{
			name: "非法邮箱",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, nil
			},
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
"email": "123qq.com",
"password": "hello#world123",
"confirmPassword": "hello#world123"
}`)))

				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: 200,
			wantBody: "邮箱格式错误",
		},
		{
			name: "两次密码不正确",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, nil
			},
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
"email": "123@qq.com",
"password": "hello#world1232",
"confirmPassword": "hello#world123"
}`)))

				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: 200,
			wantBody: "两次密码不正确",
		},
		{
			name: "密码格式不正确",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				return userSvc, nil
			},
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
"email": "123@qq.com",
"password": "h1",
"confirmPassword": "h1"
}`)))

				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: 200,
			wantBody: "密码必须包含数字、特殊字符、并且长度不能小于 8 位",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(service.ErrDuplicateEmail)
				return userSvc, nil
			},
			req: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/users/signup", bytes.NewReader([]byte(`{
"email": "123@qq.com",
"password": "hello#world123",
"confirmPassword": "hello#world123"
}`)))

				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: 200,
			wantBody: "邮箱冲突，请换一个",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 构造handler
			userSvc, codeSvc := tc.mock(ctrl)
			hdl := NewUserHandler(userSvc, codeSvc)

			//启动服务,注册路由
			server := gin.Default()
			hdl.RegisterRoutes(server)

			// 准备req和resp

			req := tc.req(t)
			resp := httptest.NewRecorder()

			// 执行
			server.ServeHTTP(resp, req)

			// 断言结果
			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())

		})
	}
}
