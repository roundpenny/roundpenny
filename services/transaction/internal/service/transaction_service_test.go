// Copyright (c) 2026 RoundPenny. All rights reserved.

package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/transaction/internal/repository"
)

type mockTxRepo struct {
	createFn    func(ctx context.Context, tx *repository.Transaction) error
	getByIDFn   func(ctx context.Context, id uuid.UUID) (*repository.Transaction, error)
	listByUserFn func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]repository.Transaction, error)
	settleFn    func(ctx context.Context, id uuid.UUID) error
}

func (m *mockTxRepo) Create(ctx context.Context, tx *repository.Transaction) error {
	return m.createFn(ctx, tx)
}
func (m *mockTxRepo) GetByID(ctx context.Context, id uuid.UUID) (*repository.Transaction, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockTxRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]repository.Transaction, error) {
	return m.listByUserFn(ctx, userID, limit, offset)
}
func (m *mockTxRepo) Settle(ctx context.Context, id uuid.UUID) error {
	return m.settleFn(ctx, id)
}

type mockProducer struct {
	publishFn func(ctx context.Context, msg kafka.Message) error
}

func (m *mockProducer) Publish(ctx context.Context, msg kafka.Message) error {
	return m.publishFn(ctx, msg)
}
func (m *mockProducer) Close() error { return nil }

func TestCreateTransaction_Success(t *testing.T) {
	txID := uuid.New()
	userID := uuid.New()

	var published bool
	svc := NewTransactionService(&mockTxRepo{
		createFn: func(ctx context.Context, tx *repository.Transaction) error {
			tx.ID = txID
			tx.Status = "pending"
			tx.CreatedAt = time.Now()
			return nil
		},
		settleFn: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Transaction, error) {
			return &repository.Transaction{
				ID:     id,
				UserID: userID,
				Amount: 24.53,
				Status: "settled",
			}, nil
		},
	}, &mockProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			published = true
			if msg.Topic != event.TopicTransactionSettled {
				t.Fatalf("expected topic %s, got %s", event.TopicTransactionSettled, msg.Topic)
			}
			return nil
		},
	})

	tx, err := svc.Create(context.Background(), CreateTransactionRequest{
		UserID:   userID.String(),
		Amount:   24.53,
		Currency: "USD",
		Type:     "purchase",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.ID != txID {
		t.Fatalf("expected ID %v, got %v", txID, tx.ID)
	}
	if !published {
		t.Fatal("expected event to be published")
	}
}

func TestCreateTransaction_InvalidUserID(t *testing.T) {
	svc := NewTransactionService(&mockTxRepo{}, &mockProducer{})
	_, err := svc.Create(context.Background(), CreateTransactionRequest{
		UserID:   "not-a-uuid",
		Amount:   10,
		Currency: "USD",
		Type:     "purchase",
	})
	if err == nil {
		t.Fatal("expected error for invalid user_id")
	}
}

func TestListByUser_PaginationDefaults(t *testing.T) {
	userID := uuid.New()
	svc := NewTransactionService(&mockTxRepo{
		listByUserFn: func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]repository.Transaction, error) {
			if limit != 20 {
				t.Fatalf("expected default limit 20, got %d", limit)
			}
			if offset != 0 {
				t.Fatalf("expected offset 0 for page 1, got %d", offset)
			}
			return []repository.Transaction{}, nil
		},
	}, &mockProducer{})

	txs, err := svc.ListByUser(context.Background(), userID, 0, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(txs) != 0 {
		t.Fatalf("expected empty list")
	}
}

func TestListByUser_Page2(t *testing.T) {
	userID := uuid.New()
	svc := NewTransactionService(&mockTxRepo{
		listByUserFn: func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]repository.Transaction, error) {
			if limit != 10 {
				t.Fatalf("expected limit 10, got %d", limit)
			}
			if offset != 10 {
				t.Fatalf("expected offset 10 for page 2, got %d", offset)
			}
			return []repository.Transaction{}, nil
		},
	}, &mockProducer{})

	txs, err := svc.ListByUser(context.Background(), userID, 2, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(txs) != 0 {
		t.Fatalf("expected empty list")
	}
}
