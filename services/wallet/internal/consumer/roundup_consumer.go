package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/services/wallet/internal/service"
)

type RoundUpConsumer struct {
	svc *service.WalletService
}

func NewRoundUpConsumer(svc *service.WalletService) *RoundUpConsumer {
	return &RoundUpConsumer{svc: svc}
}

func (c *RoundUpConsumer) HandleRoundUp(ctx context.Context, topic string, key string, data []byte) error {
	var evt event.RoundUpCalculated
	if err := json.Unmarshal(data, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		return fmt.Errorf("parse user_id: %w", err)
	}

	roundUpID, err := uuid.Parse(evt.TransactionID)
	if err != nil {
		return fmt.Errorf("parse tx_id: %w", err)
	}

	entry, err := c.svc.CreditRoundUp(ctx, userID, evt.RoundUpAmount, roundUpID)
	if err != nil {
		return fmt.Errorf("credit: %w", err)
	}

	slog.Info("wallet credited", "user", evt.UserID, "amount", evt.RoundUpAmount, "balance", entry.BalanceAfter)

	return nil
}
