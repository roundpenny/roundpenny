// Copyright (c) 2026 RoundPenny. All rights reserved.

package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/fee/internal/repository"
)

type FeeRepository interface {
	GetActiveConfig(ctx context.Context, feeType string) (*repository.FeeConfig, error)
	CreateFeeTransaction(ctx context.Context, ft *repository.FeeTransaction) error
}

type FeeProducer interface {
	Publish(ctx context.Context, msg kafka.Message) error
}

type FeeService struct {
	repo     FeeRepository
	producer FeeProducer
}

func NewFeeService(repo FeeRepository, producer FeeProducer) *FeeService {
	return &FeeService{repo: repo, producer: producer}
}

func (s *FeeService) ChargeRoundUpFee(ctx context.Context, roundUpID uuid.UUID, userID uuid.UUID, roundUpAmount float64) error {
	cfg, err := s.repo.GetActiveConfig(ctx, "roundup")
	if err != nil {
		slog.Info("no fee config found, using default 10%")
		defaultPct := 10.00
		cfg = &repository.FeeConfig{
			FeeType:    "roundup",
			Percentage: &defaultPct,
		}
	}

	var feeAmount float64
	if cfg.Percentage != nil {
		feeAmount = roundUpAmount * (*cfg.Percentage) / 100.0
	} else if cfg.FlatAmount != nil {
		feeAmount = *cfg.FlatAmount
	}

	if cfg.MinAmount != nil && feeAmount < *cfg.MinAmount {
		feeAmount = *cfg.MinAmount
	}
	if cfg.MaxAmount != nil && feeAmount > *cfg.MaxAmount {
		feeAmount = *cfg.MaxAmount
	}

	ft := &repository.FeeTransaction{
		UserID:           userID,
		Amount:           feeAmount,
		FeeType:          "roundup",
		PercentageApplied: cfg.Percentage,
		Currency:         "USD",
		Status:           "charged",
	}

	if err := s.repo.CreateFeeTransaction(ctx, ft); err != nil {
		return fmt.Errorf("create fee: %w", err)
	}

	evt := event.FeeCharged{
		TransactionID: roundUpID.String(),
		UserID:        userID.String(),
		Amount:        feeAmount,
		FeeType:       "roundup",
	}

	if err := s.producer.Publish(ctx, kafka.Message{
		Key:     userID.String(),
		Topic:   event.TopicFeeCharged,
		Payload: evt,
	}); err != nil {
		slog.Warn("fee publish warning", "error", err)
	}

	if cfg.Percentage != nil {
		slog.Info("fee charged", "user", userID, "amount", feeAmount, "percentage", *cfg.Percentage, "total", roundUpAmount)
	} else {
		slog.Info("fee charged", "user", userID, "amount", feeAmount, "type", "flat")
	}

	return nil
}
