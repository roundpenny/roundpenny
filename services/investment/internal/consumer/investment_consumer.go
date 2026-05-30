// Copyright (c) 2026 RoundPenny. All rights reserved.

package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/services/investment/internal/service"
)

type InvestmentConsumer struct {
	svc *service.InvestmentService
}

func NewInvestmentConsumer(svc *service.InvestmentService) *InvestmentConsumer {
	return &InvestmentConsumer{svc: svc}
}

func (c *InvestmentConsumer) HandleRoundUp(ctx context.Context, topic string, key string, data []byte) error {
	var evt event.RoundUpCalculated
	if err := json.Unmarshal(data, &evt); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		return fmt.Errorf("parse user_id: %w", err)
	}

	if err := c.svc.InvestRoundUp(ctx, userID, evt.RoundUpAmount); err != nil {
		slog.Error("invest failed", "error", err)
		return nil
	}

	return nil
}
