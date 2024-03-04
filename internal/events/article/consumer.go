package article

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/internal/repository"
	"webook/pkg/logger"
	"webook/pkg/saramax"
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

	// 异步执行消费逻辑
	go func() {
		er := cg.Consume(context.Background(), []string{TopicReadEvent}, saramax.NewHandler[ReadEvent](i.l, i.Consume))

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
