package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

type Handler func(ctx context.Context, topic string, key string, data []byte) error

func NewConsumer(groupID string, topics []string) (*Consumer, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:  []string{brokers},
		GroupID:  groupID,
		MinBytes: 10,
		MaxBytes: 10e6,
	}
	if len(topics) == 1 {
		readerConfig.Topic = topics[0]
	} else {
		readerConfig.GroupTopics = topics
	}

	if os.Getenv("KAFKA_TLS_ENABLED") == "true" {
		dialer := &kafka.Dialer{
			Timeout: 10 * time.Second,
			TLS:     &tls.Config{},
		}
		readerConfig.Dialer = dialer
	}

	reader := kafka.NewReader(readerConfig)

	return &Consumer{reader: reader}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler Handler) error {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("read message: %w", err)
		}

		if err := handler(ctx, msg.Topic, string(msg.Key), msg.Value); err != nil {
			return fmt.Errorf("handle message: %w", err)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func UnmarshalPayload[T any](data []byte) (*T, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &v, nil
}
