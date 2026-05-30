// Copyright (c) 2026 RoundPenny. All rights reserved.

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

type Producer struct {
	writer *kafka.Writer
}

type Message struct {
	Key     string
	Topic   string
	Payload any
}

func NewProducer() (*Producer, error) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers),
		Balancer:     &kafka.Hash{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireAll,
	}

	if os.Getenv("KAFKA_TLS_ENABLED") == "true" {
		tlsCfg := &tls.Config{}
		writer.Transport = &kafka.Transport{
			TLS: tlsCfg,
		}
	}

	return &Producer{writer: writer}, nil
}

func (p *Producer) Publish(ctx context.Context, msg Message) error {
	data, err := json.Marshal(msg.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(msg.Key),
		Topic: msg.Topic,
		Value: data,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
