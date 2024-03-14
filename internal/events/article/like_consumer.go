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

type InteractiveLikeEventConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
	l      logger.Logger
	group  string
}

func NewInteractiveLikeEventConsumer(repo repository.InteractiveRepository, client sarama.Client, l logger.Logger) *InteractiveLikeEventConsumer {
	return &InteractiveLikeEventConsumer{repo: repo, client: client, l: l, group: "interactive"}
}

func (i *InteractiveLikeEventConsumer) Start() error {
	// 创建consumer
	cg, err := sarama.NewConsumerGroupFromClient(i.group, i.client)

	if err != nil {
		return err
	}

	opts := prometheus.SummaryOpts{
		Namespace: "webook_kafka_like",
		Subsystem: "webook",
		Name:      "interactive",
		Help:      "统计 interactive 点赞事件交互",
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
		er := cg.Consume(context.Background(), []string{TopicLikeEvent}, saramax.NewHandler[LikeEvent](i.l, i.Consume, opts, "article_like_event"))

		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return nil
}

// 真正的消费逻辑
func (i *InteractiveLikeEventConsumer) Consume(msg *sarama.ConsumerMessage,
	event LikeEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrLike(ctx, "article", event.Aid, event.Uid)
}
