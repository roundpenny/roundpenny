package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

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
		slog.Info("webhook consumer starting", "topics", c.topics, "group", groupID)

		handler := func(ctx context.Context, topic string, key string, data []byte) error {
			eventID := uuid.New().String()

			var rawPayload map[string]any
			if err := json.Unmarshal(data, &rawPayload); err != nil {
				slog.Error("unmarshal event payload", "error", err)
				return nil
			}

			if err := c.svc.DeliverEvent(ctx, topic, eventID, data); err != nil {
				slog.Error("deliver event", "topic", topic, "error", err)
				return nil
			}

			return nil
		}

		if err := consumer.Consume(ctx, handler); err != nil {
			slog.Error("consumer error", "error", err)
		}
	}()

	return nil
}
