package article

import (
	"context"
	"github.com/IBM/sarama"
	"time"
	"webook/internal/repository"
	"webook/pkg/logger"
	"webook/pkg/saramax"
)

type InteractiveReadEventBatchConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
	l      logger.Logger
	group  string
}

func NewInteractiveReadEventBatchConsumer(repo repository.InteractiveRepository, client sarama.Client, l logger.Logger) *InteractiveReadEventBatchConsumer {
	return &InteractiveReadEventBatchConsumer{repo: repo, client: client, l: l, group: "interactive"}
}

func (i *InteractiveReadEventBatchConsumer) Start() error {
	// 创建consumer
	cg, err := sarama.NewConsumerGroupFromClient(i.group, i.client)

	if err != nil {
		return err
	}

	// 异步执行消费逻辑
	go func() {
		er := cg.Consume(context.Background(), []string{TopicReadEvent}, saramax.NewBatchHandler[ReadEvent](i.l, i.Consume))

		if er != nil {
			i.l.Error("退出消费", logger.Error(er))
		}
	}()
	return nil
}

// 真正的消费逻辑
func (i *InteractiveReadEventBatchConsumer) Consume(msg []*sarama.ConsumerMessage,
	events []ReadEvent) error {
	bizs := make([]string, 0, len(events))
	bizIds := make([]int64, 0, len(events))
	for _, evt := range events {
		bizs = append(bizs, "article")
		bizIds = append(bizIds, evt.Aid)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	i.l.Info("我开始异步消费处理了，这一批数据有：", logger.Int64("数据量", int64(len(msg))))
	return i.repo.BatchIncrReadCnt(ctx, bizs, bizIds)

}
