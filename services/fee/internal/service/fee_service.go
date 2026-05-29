package service

import (
	"context"
	"fmt"
	"log"

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
		log.Printf("no fee config found, using default 10%%")
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
		log.Printf("fee publish warning: %v", err)
	}

	if cfg.Percentage != nil {
		log.Printf("fee charged: user=%s amount=%.4f (%.0f%% of %.2f)",
			userID, feeAmount, *cfg.Percentage, roundUpAmount)
	} else {
		log.Printf("fee charged: user=%s amount=%.4f (flat fee)",
			userID, feeAmount)
	}

	return nil
}
