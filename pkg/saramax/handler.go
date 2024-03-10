package saramax

import (
	"encoding/json"
	"errors"
	"time"
	"webook/pkg/logger"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
)

type Handler[T any] struct {
	l      logger.Logger
	vector *prometheus.SummaryVec
	fn     func(msg *sarama.ConsumerMessage, event T) error
	event  string
}

// options
func InitSummaryVec(opts prometheus.SummaryOpts) *prometheus.SummaryVec {
	vector := prometheus.NewSummaryVec(opts, []string{"error", "event"})
	prometheus.MustRegister(vector)
	return vector
}

func NewHandler[T any](l logger.Logger, fn func(msg *sarama.ConsumerMessage, event T) error, opts prometheus.SummaryOpts, event string) *Handler[T] {
	return &Handler[T]{
		l:      l,
		fn:     fn,
		vector: InitSummaryVec(opts),
		event:  event,
	}
}

func (h *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()

	for msg := range msgs {
		// 在这里处理业务调用逻辑
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			// 记录日志
			// 你也可以在这里引入重试的逻辑
			h.l.Error("反序列消息体失败",
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset),
				logger.Error(err))
		}

		// kafka添加监控
		start := time.Now()
		err = h.fn(msg, t)

		// 暂时同步
		duration := time.Since(start).Milliseconds()
		if err == nil {
			err = errors.New("no error")
			h.vector.WithLabelValues(err.Error(), h.event).Observe(float64(duration))
			err = nil
		}

		if err != nil {
			h.l.Error("处理消息失败",
				logger.String("topic", msg.Topic),
				logger.Int32("partition", msg.Partition),
				logger.Int64("offset", msg.Offset),
				logger.Error(err))
		}

		session.MarkMessage(msg, "")
	}

	return nil
}
