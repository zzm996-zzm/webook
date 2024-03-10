package article

import (
	"context"
	"time"
	"webook/internal/repository"
	"webook/pkg/logger"
	"webook/pkg/saramax"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
)

type InteractiveReadEventConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
	l      logger.Logger
	group  string
}

func NewInteractiveReadEventConsumer(repo repository.InteractiveRepository, client sarama.Client, l logger.Logger) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{repo: repo, client: client, l: l, group: "interactive"}
}

func (i *InteractiveReadEventConsumer) Start() error {
	// 创建consumer
	cg, err := sarama.NewConsumerGroupFromClient(i.group, i.client)

	if err != nil {
		return err
	}

	opts := prometheus.SummaryOpts{
		Namespace: "webook_kafka",
		Subsystem: "webook",
		Name:      "interactive",
		Help:      "统计 interactive 事件交互",
		ConstLabels: map[string]string{
			"instance_id": "my_kafka",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}

	// 异步执行消费逻辑
	go func() {
		er := cg.Consume(context.Background(), []string{TopicReadEvent}, saramax.NewHandler[ReadEvent](i.l, i.Consume, opts, "article_read_event"))

		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return nil
}

// 真正的消费逻辑
func (i *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage,
	event ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrReadCnt(ctx, "article", event.Aid)
}
