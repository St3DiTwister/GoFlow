package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"time"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3,
			MaxBytes: 10e6,
			MaxWait:  100 * time.Millisecond,
		}),
	}
}

func (c *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
