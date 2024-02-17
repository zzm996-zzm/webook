package repository

import (
	"context"
	"webook/internal/domain"
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
