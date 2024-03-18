package dao

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

//go:generate mockgen -source=./async_sms.go -package=daomocks -destination=mocks/async_sms.mock.go AsyncSmsDAO
type AsyncSmsDAO interface {
	Insert(ctx context.Context, s AsyncSms) error
	GetWaitingSMS(ctx context.Context) (AsyncSms, error)
	MarkSuccess(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
}

const (
	// 等待
	asyncStatusWaiting = iota
	// 失败
	asyncStatusFailed
	// 成功
	asyncStatusSuccess
)

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

type GORMAsyncSmsDAO struct {
	db *gorm.DB
}

func NewGORMAsyncSmsDAO(db *gorm.DB) AsyncSmsDAO {
	return &GORMAsyncSmsDAO{db: db}
}

// Insert 插入一条需要异步执行的sms
func (a *GORMAsyncSmsDAO) Insert(ctx context.Context, s AsyncSms) error {
	return a.db.Create(&s).Error
}

func (a *GORMAsyncSmsDAO) GetWaitingSMS(ctx context.Context) (AsyncSms, error) {
	// 如果在高并发情况下, select for update 对数据库压力很大
	// 但是我们不是高并发，部署N台机器，才有N个goroutine 来查询
	// 假设并发不过百那就是随便写
	var s AsyncSms
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		endTime := now - time.Minute.Milliseconds()
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("utime > ? and status = ?", endTime, asyncStatusWaiting).First(&s).Error

		if err != nil {
			return err
		}

		// 查询出等待的sms记录，并且重试次数++
		err = tx.Model(&AsyncSms{}).Where("id = ?", s.Id).Updates(map[string]any{
			"retry_cnt": gorm.Expr("retry_cnt + 1"),
			"utime":     now,
		}).Error

		return err
	})
	return s, err
}

// 标记发送成功的sms
func (a *GORMAsyncSmsDAO) MarkSuccess(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return a.db.WithContext(ctx).Model(&AsyncSms{}).Where("id = ? ", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusSuccess,
		}).Error
}
func (a *GORMAsyncSmsDAO) MarkFailed(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return a.db.WithContext(ctx).Model(&AsyncSms{}).
		// 只有到达了重试次数才会更新
		Where("id =? and `retry_cnt`>=`retry_max`", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusFailed,
		}).Error

}

type SmsConfig struct {
	TplId   string
	Args    []string
	Numbers []string
}
