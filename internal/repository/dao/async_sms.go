package dao

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
)

//go:generate mockgen -source=./async_sms.go -package=daomocks -destination=mocks/async_sms.mock.go AsyncSmsDAO
type AsyncSmsDAO interface {
	Insert(ctx context.Context, s AsyncSms) error
	GetWaitingSMS(ctx context.Context) (AsyncSms, error)
	MarkSuccess(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
}

type AsyncSms struct {
	Id int64
	// 使用我在 ekit 里面支持的 JSON 字段
	Config sqlx.JsonColumn[SmsConfig]
	// 重试次数
	RetryCnt int
	// 重试的最大次数
	RetryMax int
	Status   uint8
	CTime    int64
	UTime    int64 `gorm:"index"`
}

type SmsConfig struct {
	TplId   string
	Args    []string
	Numbers []string
}
