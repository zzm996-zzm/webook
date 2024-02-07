package service

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"webook/internal/domain"
	"webook/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户名或者密码不对")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) Signup(ctx context.Context, u domain.User) error {
	// 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)

	if err != nil {
		return err
	}

	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}

	// 检查密码对不对
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))

	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}

	return u, nil
}

func (svc *UserService) UpdateNonSensitiveInfo(ctx *gin.Context, u domain.User) error {
	return svc.repo.UpdateNonZeroFields(ctx, u)
}

func (svc *UserService) FindById(ctx *gin.Context, uid int64) (domain.User, error) {
	return svc.repo.FindById(ctx, uid)
}

func (svc *UserService) FindOrCreate(ctx *gin.Context, phone string) (domain.User, error) {
	u, err := svc.repo.FindByPhone(ctx, phone)
	if !errors.Is(err, repository.ErrUserNotFound) {
		// 有两种情况
		// err == nil, u 是可用的
		// err != nil，系统错误，
		return u, err
	}
	err = svc.repo.Create(ctx, domain.User{Phone: phone})

	// 有两种可能，一种是 err 恰好是唯一索引冲突（phone）
	// 一种是 err != nil，系统错误
	if err != nil && err != repository.ErrDuplicateUser {
		return domain.User{}, err
	}
	// 要么 err ==nil，要么ErrDuplicateUser，也代表用户存在
	// 主从延迟，理论上来讲，强制走主库
	return svc.repo.FindByPhone(ctx, phone)
}
