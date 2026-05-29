package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/services/fee/internal/service"
)

type FeeConsumer struct {
	svc *service.FeeService
}

func NewFeeConsumer(svc *service.FeeService) *FeeConsumer {
	return &FeeConsumer{svc: svc}
}

func (c *FeeConsumer) HandleRoundUp(ctx context.Context, topic string, key string, data []byte) error {
	var evt event.RoundUpCalculated
	if err := json.Unmarshal(data, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		return fmt.Errorf("parse user_id: %w", err)
	}

	txID, err := uuid.Parse(evt.TransactionID)
	if err != nil {
		return fmt.Errorf("parse tx_id: %w", err)
	}

	if err := c.svc.ChargeRoundUpFee(ctx, txID, userID, evt.RoundUpAmount); err != nil {
		log.Printf("fee charge failed: %v", err)
		return nil
	}

	return nil
}
