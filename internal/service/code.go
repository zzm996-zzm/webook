package service

import (
	"context"
	"fmt"
	"math/rand"
	"webook/internal/repository"
	"webook/internal/service/sms"
)

var ErrCodeSendTooMany = repository.ErrCodeSendTooMany
var ErrCodeVerifyTooMany = repository.ErrCodeVerifyTooMany

type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context,
		biz, phone, inputCode string) (bool, error)
}
type codeService struct {
	repo repository.CodeRepository
	sms  sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &codeService{
		repo: repo,
		sms:  smsSvc,
	}
}

func (c *codeService) Send(ctx context.Context, biz, phone string) error {
	code := c.generate()
	err := c.repo.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}

	tmpId := "577583"
	expiration := "10"
	return c.sms.Send(ctx, tmpId, []string{code, expiration}, phone)

}

func (c *codeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	//验证redis是否有数据
	ok, err := c.repo.Verify(ctx, biz, phone, inputCode)
	if err != nil {
		return false, nil
	}

	return ok, nil
}

func (c *codeService) generate() string {
	// 0-999999
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
