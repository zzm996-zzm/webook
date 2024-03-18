package repository

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"webook/internal/domain"
	"webook/internal/repository/dao"
)

type AsyncSmsRepository interface {
	// Add 添加一个异步 SMS 记录。
	// 你叫做 Create 或者 Insert 也可以
	Add(ctx context.Context, s domain.AsyncSms) error
	// PreemptWaitingSMS 抢占等待发生的短信
	PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error)
	// ReportScheduleResult 返回调度结果，抢占成功还是抢占失败
	ReportScheduleResult(ctx context.Context, id int64, success bool) error
}

type asyncSmsRepository struct {
	dao dao.AsyncSmsDAO
}

func (a *asyncSmsRepository) Add(ctx context.Context, s domain.AsyncSms) error {
	return a.dao.Insert(ctx, dao.AsyncSms{
		Id: s.Id,
		Config: sqlx.JsonColumn[dao.SmsConfig]{
			Val: dao.SmsConfig{
				TplId:   s.TplId,
				Args:    s.Args,
				Numbers: s.Numbers,
			},
			Valid: true,
		},
		RetryMax: s.RetryMax,
	})
}

func (a *asyncSmsRepository) PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error) {
	as, err := a.dao.GetWaitingSMS(ctx)
	if err != nil {
		return domain.AsyncSms{}, err
	}

	return domain.AsyncSms{
		Id:       as.Id,
		TplId:    as.Config.Val.TplId,
		Args:     as.Config.Val.Args,
		Numbers:  as.Config.Val.Numbers,
		RetryMax: as.RetryMax,
	}, nil
}

// ReportScheduleResult service 传递 success 报告调度结果，如果成功则标记成功，否则标记失败
func (a *asyncSmsRepository) ReportScheduleResult(ctx context.Context, id int64, success bool) error {
	if success {
		return a.dao.MarkSuccess(ctx, id)
	}
	return a.dao.MarkFailed(ctx, id)
}

func NewAsyncSmSRepository(dao dao.AsyncSmsDAO) AsyncSmsRepository {
	return &asyncSmsRepository{dao: dao}
}
