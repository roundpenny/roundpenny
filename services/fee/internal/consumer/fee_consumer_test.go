// Copyright (c) 2026 RoundPenny. All rights reserved.

package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/fee/internal/repository"
	"github.com/roundup-platform/services/fee/internal/service"
)

type mockFeeRepo struct {
	getActiveConfigFn       func(ctx context.Context, feeType string) (*repository.FeeConfig, error)
	createFeeTransactionFn func(ctx context.Context, ft *repository.FeeTransaction) error
}

func (m *mockFeeRepo) GetActiveConfig(ctx context.Context, feeType string) (*repository.FeeConfig, error) {
	return m.getActiveConfigFn(ctx, feeType)
}
func (m *mockFeeRepo) CreateFeeTransaction(ctx context.Context, ft *repository.FeeTransaction) error {
	return m.createFeeTransactionFn(ctx, ft)
}

type mockFeeProd struct {
	publishFn func(ctx context.Context, msg kafka.Message) error
}

func (m *mockFeeProd) Publish(ctx context.Context, msg kafka.Message) error {
	return m.publishFn(ctx, msg)
}
func (m *mockFeeProd) Close() error { return nil }

func TestFeeConsumer_HandleRoundUp_Success(t *testing.T) {
	var called bool
	svc := service.NewFeeService(&mockFeeRepo{
		getActiveConfigFn: func(ctx context.Context, feeType string) (*repository.FeeConfig, error) {
			pct := 10.00
			return &repository.FeeConfig{FeeType: "roundup", Percentage: &pct, IsActive: true}, nil
		},
		createFeeTransactionFn: func(ctx context.Context, ft *repository.FeeTransaction) error {
			called = true
			return nil
		},
	}, &mockFeeProd{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			return nil
		},
	})

	c := NewFeeConsumer(svc)
	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, uuid.New().String(), mustJSON(event.RoundUpCalculated{
		TransactionID: uuid.New().String(),
		UserID:        uuid.New().String(),
		RoundUpAmount: 0.88,
		Currency:      "USD",
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected CreateFeeTransaction to be called")
	}
}

func TestFeeConsumer_HandleRoundUp_InvalidJSON(t *testing.T) {
	svc := service.NewFeeService(&mockFeeRepo{}, &mockFeeProd{})
	c := NewFeeConsumer(svc)

	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, uuid.New().String(), []byte("not-json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFeeConsumer_HandleRoundUp_InvalidUUID(t *testing.T) {
	svc := service.NewFeeService(&mockFeeRepo{}, &mockFeeProd{})
	c := NewFeeConsumer(svc)

	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, uuid.New().String(), mustJSON(event.RoundUpCalculated{
		TransactionID: "not-a-uuid",
		UserID:        uuid.New().String(),
		RoundUpAmount: 0.88,
	}))
	if err == nil {
		t.Fatal("expected error for invalid transaction ID")
	}
}

func TestFeeConsumer_HandleRoundUp_DBError(t *testing.T) {
	svc := service.NewFeeService(&mockFeeRepo{
		getActiveConfigFn: func(ctx context.Context, feeType string) (*repository.FeeConfig, error) {
			pct := 10.00
			return &repository.FeeConfig{FeeType: "roundup", Percentage: &pct, IsActive: true}, nil
		},
		createFeeTransactionFn: func(ctx context.Context, ft *repository.FeeTransaction) error {
			return errors.New("db unavailable")
		},
	}, &mockFeeProd{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			return nil
		},
	})

	c := NewFeeConsumer(svc)
	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, uuid.New().String(), mustJSON(event.RoundUpCalculated{
		TransactionID: uuid.New().String(),
		UserID:        uuid.New().String(),
		RoundUpAmount: 0.88,
	}))
	if err != nil {
		t.Fatalf("expected no error (consumer swallows DB errors), got %v", err)
	}
}

func mustJSON(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}
