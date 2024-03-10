package ioc

import (
	"webook/pkg/saramax"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
)

// 监控kafka消息偏移量
func InitKafkaPrometheus(client sarama.Client) *saramax.MonitorMessage {
	opts := prometheus.GaugeOpts{
		Namespace: "kafka",
		Subsystem: "consumer",
		Name:      "lag",
		Help:      "The lag of a consumer group for a specific topic partition.",
	}

	monitorMessage := saramax.NewMonitorMessage(opts, client)

	return monitorMessage
}
