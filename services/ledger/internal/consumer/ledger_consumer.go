package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/services/ledger/internal/service"
)

type LedgerConsumer struct {
	svc *service.LedgerService
}

func NewLedgerConsumer(svc *service.LedgerService) *LedgerConsumer {
	return &LedgerConsumer{svc: svc}
}

func (c *LedgerConsumer) HandleEvent(ctx context.Context, topic string, key string, data []byte) error {
	switch topic {
	case event.TopicRoundUpCalculated:
		var evt event.RoundUpCalculated
		if err := json.Unmarshal(data, &evt); err != nil {
			return fmt.Errorf("unmarshal roundup: %w", err)
		}
		return c.svc.RecordRoundUp(ctx, evt)

	case event.TopicFeeCharged:
		var evt event.FeeCharged
		if err := json.Unmarshal(data, &evt); err != nil {
			return fmt.Errorf("unmarshal fee: %w", err)
		}
		return c.svc.RecordFee(ctx, evt)

	case event.TopicInvestmentCreated:
		var evt event.InvestmentCreated
		if err := json.Unmarshal(data, &evt); err != nil {
			return fmt.Errorf("unmarshal investment: %w", err)
		}
		return c.svc.RecordInvestment(ctx, evt)

	default:
		slog.Warn("ledger: unknown topic", "topic", topic)
		return nil
	}
}
