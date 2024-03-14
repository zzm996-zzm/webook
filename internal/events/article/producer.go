package article

import (
	"encoding/json"
	"github.com/IBM/sarama"
)

const TopicReadEvent = "article_read"
const TopicLikeEvent = "article_like"

type Producer interface {
	ProducerReadEvent(evt ReadEvent) error
	ProducerLikeEvent(evt LikeEvent) error
	ProducerUnLikeEvent(evt UnLikeEvent) error
}

type ReadEvent struct {
	Aid int64
	Uid int64
}

type LikeEvent struct {
	Aid int64
	Uid int64
}

type UnLikeEvent struct {
	Aid int64
	Uid int64
}

type SaramaSyncProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaSyncProducer(producer sarama.SyncProducer) Producer {
	return &SaramaSyncProducer{producer: producer}
}

func (s *SaramaSyncProducer) ProducerReadEvent(evt ReadEvent) error {
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	// 发送消息
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicReadEvent,
		Value: sarama.StringEncoder(val),
	})

	return err
}

func (s *SaramaSyncProducer) ProducerLikeEvent(evt LikeEvent) error {
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	// 发送消息
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicLikeEvent,
		Value: sarama.StringEncoder(val),
	})

	return err
}

func (s *SaramaSyncProducer) ProducerUnLikeEvent(evt UnLikeEvent) error {
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	// 发送消息
	_, _, err = s.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicLikeEvent,
		Value: sarama.StringEncoder(val),
	})

	return err
}
