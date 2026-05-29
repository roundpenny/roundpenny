package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/notification/internal/service"
)

type WebhookConsumer struct {
	svc    *service.WebhookService
	topics []string
}

func NewWebhookConsumer(svc *service.WebhookService) *WebhookConsumer {
	return &WebhookConsumer{
		svc:    svc,
		topics: []string{"tx.settled", "roundup.calculated", "wallet.credited", "fee.charged", "investment.created"},
	}
}

func (c *WebhookConsumer) Start(ctx context.Context, brokers string, groupID string) error {
	consumer, err := kafka.NewConsumer(groupID, c.topics)
	if err != nil {
		return err
	}

	go func() {
		log.Printf("webhook consumer starting, topics=%v, group=%s", c.topics, groupID)

		handler := func(ctx context.Context, topic string, key string, data []byte) error {
			eventID := uuid.New().String()

			var rawPayload map[string]any
			if err := json.Unmarshal(data, &rawPayload); err != nil {
				log.Printf("unmarshal event payload: %v", err)
				return nil
			}

			if err := c.svc.DeliverEvent(ctx, topic, eventID, data); err != nil {
				log.Printf("deliver event %s: %v", topic, err)
				return nil
			}

			return nil
		}

		if err := consumer.Consume(ctx, handler); err != nil {
			log.Printf("consumer error: %v", err)
		}
	}()

	return nil
}
