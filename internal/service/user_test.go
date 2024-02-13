package service

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/mock/gomock"
	"testing"
	"webook/internal/domain"
	"webook/internal/repository"
	repomocks "webook/internal/repository/mock"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordEncrypt(t *testing.T) {
	password := []byte("1234563#123456")
	//代价（cost） 越高，加密强度越高但是对cpu负载更大
	encrypted, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	assert.NoError(t, err)
	fmt.Println(string(encrypted))

	//测试解密
	err = bcrypt.CompareHashAndPassword(encrypted, []byte("testting"))
	assert.NotNil(t, err)
	err = bcrypt.CompareHashAndPassword(encrypted, password)
	assert.NoError(t, err)
}

func TestUserService_Login(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository

		// 预期输入
		ctx      context.Context
		email    string
		password string

		// 预期返回
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(
					domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$iPbohtixkOtB8NnD4Is49eZpq9/WAC1ZFBv//cMqSrVK8XXCFR.7C",
						Phone:    "17674123135",
					}, nil)

				return userRepo
			},
			ctx:      context.Background(),
			email:    "123@qq.com",
			password: "1234563#123456",

			wantUser: domain.User{
				Email:    "123@qq.com",
				Password: "$2a$10$iPbohtixkOtB8NnD4Is49eZpq9/WAC1ZFBv//cMqSrVK8XXCFR.7C",
				Phone:    "17674123135",
			},
			wantErr: nil,
		},
		{
			name: "用户查找失败",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(
					domain.User{}, repository.ErrUserNotFound)

				return userRepo
			},
			ctx:      context.Background(),
			email:    "123@qq.com",
			password: "1234563#123456",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(
					domain.User{}, errors.New("db 错误"))

				return userRepo
			},
			ctx:      context.Background(),
			email:    "123@qq.com",
			password: "1234563#123456",

			wantUser: domain.User{},
			wantErr:  errors.New("db 错误"),
		},
		{
			name: "密码不正确",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(
					domain.User{
						Email:    "123@qq.com",
						Password: "$2a$10$iPbohtixkOtB8NnD4Is49eZpq9/WAC1ZFBv//cMqSrVK8XXCFR.7C",
						Phone:    "17674123135",
					}, nil)

				return userRepo
			},
			ctx:      context.Background(),
			email:    "123@qq.com",
			password: "1234563#123456222",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := tc.mock(ctrl)
			userSvc := NewUserService(userRepo)

			user, err := userSvc.Login(context.Background(), tc.email, tc.password)
			assert.Equal(t, tc.wantUser, user)
			assert.Equal(t, tc.wantErr, err)
		})
	}

}

func TestUserService_Signup(t *testing.T) {

}
