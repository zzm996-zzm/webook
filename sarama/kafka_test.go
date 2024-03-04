package sarama

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"log"
	"testing"
	"time"
)

var addr = []string{"localhost:9094"}

func TestProducer(t *testing.T) {
	// 直接初始化默认配置
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(addr, cfg)
	assert.NoError(t, err)
	for i := 0; i < 100; i++ {
		_, _, err = producer.SendMessage(&sarama.ProducerMessage{
			Topic: "test_topic",
			Value: sarama.StringEncoder("这是一条消息"),
			Headers: []sarama.RecordHeader{
				{
					Key:   []byte("key1"),
					Value: []byte("value1"),
				},
			},
			Metadata: "这是 metadata",
		})
	}

}

func TestAsyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(addr, cfg)
	assert.NoError(t, err)
	msgs := producer.Input()
	for i := 0; i < 100; i++ {
		msgs <- &sarama.ProducerMessage{
			Topic: "test_topic",
			Value: sarama.StringEncoder("这是一条消息"),
			// 会在生产者和消费者之间传递的
			Headers: []sarama.RecordHeader{
				{
					Key:   []byte("key1"),
					Value: []byte("value1"),
				},
			},
			Metadata: "这是 metadata",
		}

		select {
		case msg := <-producer.Successes():
			t.Log("发送成功", string(msg.Value.(sarama.StringEncoder)))
		case err := <-producer.Errors():
			t.Log("发送失败", err.Err, err.Msg)
		}
	}
}

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	consumer, err := sarama.NewConsumerGroup(addr, "demo", cfg)
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	go func() {
		time.Sleep(time.Second * 10)
		consumer.Close()
	}()
	err = consumer.Consume(ctx, []string{"test_topic"}, ConsumerHandler{})
	if err != nil {
		return
	}
}

type ConsumerHandler struct {
}

func (c ConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Println("这是Setup")
	return nil
}

func (c ConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("这是Cleanup")
	return nil
}

func (c ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	// 真正消费逻辑
	msgs := claim.Messages()
	const batchSize = 10
	// 异步消费 批量提交
	for {
		log.Println("一个批次开始")
		batch := make([]*sarama.ConsumerMessage, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var done bool
		var eg errgroup.Group
		for i := 0; i < 10 && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					return nil
				}
				batch = append(batch, msg)
				eg.Go(func() error {
					// 消费消费
					log.Println(string(msg.Value))
					return nil
				})
			}
		}

		cancel()
		// 全部处理完成
		err := eg.Wait()
		if err != nil {
			log.Println(err)
			continue
		}

		// 凑够了一批

		for _, msg := range batch {
			session.MarkMessage(msg, "")
		}

	}

}
