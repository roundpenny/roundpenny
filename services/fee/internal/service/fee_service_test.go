// Copyright (c) 2026 RoundPenny. All rights reserved.

package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/fee/internal/repository"
)

type mockFeeRepo struct {
	getActiveConfigFn     func(ctx context.Context, feeType string) (*repository.FeeConfig, error)
	createFeeTransactionFn func(ctx context.Context, ft *repository.FeeTransaction) error
}

func (m *mockFeeRepo) GetActiveConfig(ctx context.Context, feeType string) (*repository.FeeConfig, error) {
	return m.getActiveConfigFn(ctx, feeType)
}
func (m *mockFeeRepo) CreateFeeTransaction(ctx context.Context, ft *repository.FeeTransaction) error {
	return m.createFeeTransactionFn(ctx, ft)
}

type mockFeeProducer struct {
	publishFn func(ctx context.Context, msg kafka.Message) error
}

func (m *mockFeeProducer) Publish(ctx context.Context, msg kafka.Message) error {
	return m.publishFn(ctx, msg)
}
func (m *mockFeeProducer) Close() error { return nil }

func almostEqual(a, b, eps float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < eps
}

func TestChargeRoundUpFee_Percentage(t *testing.T) {
	tenPct := 10.00

	var published bool
	svc := NewFeeService(&mockFeeRepo{
		getActiveConfigFn: func(ctx context.Context, feeType string) (*repository.FeeConfig, error) {
			return &repository.FeeConfig{
				FeeType:    "roundup",
				Percentage: &tenPct,
				IsActive:   true,
			}, nil
		},
		createFeeTransactionFn: func(ctx context.Context, ft *repository.FeeTransaction) error {
			if !almostEqual(ft.Amount, 0.088, 0.0001) {
				t.Fatalf("expected fee ~0.088 (10%% of 0.88), got %.4f", ft.Amount)
			}
			if ft.FeeType != "roundup" {
				t.Fatalf("expected fee type roundup, got %s", ft.FeeType)
			}
			return nil
		},
	}, &mockFeeProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			published = true
			if msg.Topic != event.TopicFeeCharged {
				t.Fatalf("expected topic %s, got %s", event.TopicFeeCharged, msg.Topic)
			}
			return nil
		},
	})

	err := svc.ChargeRoundUpFee(context.Background(), uuid.New(), uuid.New(), 0.88)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !published {
		t.Fatal("expected event to be published")
	}
}

func TestChargeRoundUpFee_DefaultFallback(t *testing.T) {
	var published bool
	svc := NewFeeService(&mockFeeRepo{
		getActiveConfigFn: func(ctx context.Context, feeType string) (*repository.FeeConfig, error) {
			return nil, repository.ErrNotFound
		},
		createFeeTransactionFn: func(ctx context.Context, ft *repository.FeeTransaction) error {
			if !almostEqual(ft.Amount, 0.088, 0.0001) {
				t.Fatalf("expected fee ~0.088 (default 10%% of 0.88), got %.4f", ft.Amount)
			}
			return nil
		},
	}, &mockFeeProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			published = true
			return nil
		},
	})

	err := svc.ChargeRoundUpFee(context.Background(), uuid.New(), uuid.New(), 0.88)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !published {
		t.Fatal("expected event to be published")
	}
}

func TestChargeRoundUpFee_WithMinMax(t *testing.T) {
	fivePct := 5.00
	minFee := 0.05
	maxFee := 0.50

	svc := NewFeeService(&mockFeeRepo{
		getActiveConfigFn: func(ctx context.Context, feeType string) (*repository.FeeConfig, error) {
			return &repository.FeeConfig{
				FeeType:    "roundup",
				Percentage: &fivePct,
				MinAmount:  &minFee,
				MaxAmount:  &maxFee,
				IsActive:   true,
			}, nil
		},
		createFeeTransactionFn: func(ctx context.Context, ft *repository.FeeTransaction) error {
			if ft.Amount != 0.05 {
				t.Fatalf("expected fee 0.05 (clamped to min), got %.4f", ft.Amount)
			}
			return nil
		},
	}, &mockFeeProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			return nil
		},
	})

	err := svc.ChargeRoundUpFee(context.Background(), uuid.New(), uuid.New(), 0.50)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestChargeRoundUpFee_FlatAmount(t *testing.T) {
	flatFee := 0.25

	svc := NewFeeService(&mockFeeRepo{
		getActiveConfigFn: func(ctx context.Context, feeType string) (*repository.FeeConfig, error) {
			return &repository.FeeConfig{
				FeeType:    "roundup",
				FlatAmount: &flatFee,
				IsActive:   true,
			}, nil
		},
		createFeeTransactionFn: func(ctx context.Context, ft *repository.FeeTransaction) error {
			if ft.Amount != 0.25 {
				t.Fatalf("expected flat fee 0.25, got %.4f", ft.Amount)
			}
			return nil
		},
	}, &mockFeeProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			return nil
		},
	})

	err := svc.ChargeRoundUpFee(context.Background(), uuid.New(), uuid.New(), 1.00)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
