package saramax

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
)

type MonitorMessage struct {
	client sarama.Client
	gauge  *prometheus.GaugeVec
}

func NewMonitorMessage(opts prometheus.GaugeOpts, client sarama.Client) *MonitorMessage {
	return &MonitorMessage{
		gauge:  prometheus.NewGaugeVec(opts, []string{"topic", "partition"}),
		client: client,
	}
}

func (m *MonitorMessage) Register() {
	offsets, err := m.GetOffsetsForAllPartitions()
	if err != nil {
		panic(err)
	}

	for topic, partitionOffsets := range offsets {
		for partition, offset := range partitionOffsets {
			m.gauge.WithLabelValues(topic, strconv.Itoa(int(partition))).Set(float64(offset))
		}
	}

	prometheus.MustRegister(m.gauge)
}

func (m *MonitorMessage) GetOffsetsForAllPartitions() (map[string]map[int32]int64, error) {

	// offset key:topic  value:map[partition]offset
	offsets := make(map[string]map[int32]int64)

	topics, err := m.client.Topics()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(len(topics))

	for _, topic := range topics {
		go func(t string) {
			defer wg.Done()

			// 获得所有的该topic下所有的patition
			partitions, err := m.client.Partitions(t)
			if err != nil {
				fmt.Println("Error getting partitions for topic", t, ":", err)
				return
			}

			topicOffsets := make(map[int32]int64)
			for _, p := range partitions {
				offset, err := m.client.GetOffset(t, p, sarama.OffsetNewest)
				if err != nil {
					fmt.Println("Error getting offset for topic:", t, ", partition:", p, ":", err)
					continue
				}
				topicOffsets[p] = offset
			}
			offsets[t] = topicOffsets
		}(topic)
	}

	wg.Wait()

	return offsets, nil
}
