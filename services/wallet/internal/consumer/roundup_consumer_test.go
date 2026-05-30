package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/services/wallet/internal/repository"
	"github.com/roundup-platform/services/wallet/internal/service"
)

type mockWalletRepo struct {
	getOrCreateFn func(ctx context.Context, userID uuid.UUID) (*repository.Wallet, error)
	creditFn      func(ctx context.Context, walletID uuid.UUID, amount float64, entryType, refType, description string, refID *uuid.UUID) (*repository.WalletEntry, error)
	getEntriesFn  func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]repository.WalletEntry, error)
}

func (m *mockWalletRepo) GetOrCreate(ctx context.Context, userID uuid.UUID) (*repository.Wallet, error) {
	return m.getOrCreateFn(ctx, userID)
}
func (m *mockWalletRepo) Credit(ctx context.Context, walletID uuid.UUID, amount float64, entryType, refType, description string, refID *uuid.UUID) (*repository.WalletEntry, error) {
	return m.creditFn(ctx, walletID, amount, entryType, refType, description, refID)
}
func (m *mockWalletRepo) GetEntries(ctx context.Context, userID uuid.UUID, limit, offset int) ([]repository.WalletEntry, error) {
	return m.getEntriesFn(ctx, userID, limit, offset)
}

func TestRoundUpConsumer_HandleRoundUp_Success(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()
	roundUpID := uuid.New()

	var credited bool
	svc := service.NewWalletService(&mockWalletRepo{
		getOrCreateFn: func(ctx context.Context, uid uuid.UUID) (*repository.Wallet, error) {
			return &repository.Wallet{ID: walletID, UserID: uid, Balance: 10, Currency: "USD", Status: "active"}, nil
		},
		creditFn: func(ctx context.Context, wid uuid.UUID, amount float64, entryType, refType, description string, refID *uuid.UUID) (*repository.WalletEntry, error) {
			credited = true
			return &repository.WalletEntry{
				WalletID:      wid,
				UserID:        userID,
				Amount:        amount,
				BalanceBefore: 10,
				BalanceAfter:  10.88,
				EntryType:     "credit",
				ReferenceType: "roundup",
				ReferenceID:   refID,
			}, nil
		},
	}, nil)

	c := NewRoundUpConsumer(svc)
	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, userID.String(), mustJSON(event.RoundUpCalculated{
		TransactionID: roundUpID.String(),
		UserID:        userID.String(),
		RoundUpAmount: 0.88,
		Currency:      "USD",
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !credited {
		t.Fatal("expected wallet to be credited")
	}
}

func TestRoundUpConsumer_HandleRoundUp_InvalidJSON(t *testing.T) {
	svc := service.NewWalletService(&mockWalletRepo{}, nil)
	c := NewRoundUpConsumer(svc)

	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, uuid.New().String(), []byte("not-json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestRoundUpConsumer_HandleRoundUp_InvalidUserID(t *testing.T) {
	svc := service.NewWalletService(&mockWalletRepo{}, nil)
	c := NewRoundUpConsumer(svc)

	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, uuid.New().String(), mustJSON(event.RoundUpCalculated{
		TransactionID: uuid.New().String(),
		UserID:        "not-a-uuid",
		RoundUpAmount: 0.88,
	}))
	if err == nil {
		t.Fatal("expected error for invalid user ID")
	}
}

func TestRoundUpConsumer_HandleRoundUp_GetOrCreateError(t *testing.T) {
	userID := uuid.New()

	svc := service.NewWalletService(&mockWalletRepo{
		getOrCreateFn: func(ctx context.Context, uid uuid.UUID) (*repository.Wallet, error) {
			return nil, errors.New("wallet not found")
		},
	}, nil)

	c := NewRoundUpConsumer(svc)
	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, userID.String(), mustJSON(event.RoundUpCalculated{
		TransactionID: uuid.New().String(),
		UserID:        userID.String(),
		RoundUpAmount: 0.88,
	}))
	if err == nil {
		t.Fatal("expected error for wallet retrieval failure")
	}
}

func TestRoundUpConsumer_HandleRoundUp_CreditError(t *testing.T) {
	userID := uuid.New()
	walletID := uuid.New()

	svc := service.NewWalletService(&mockWalletRepo{
		getOrCreateFn: func(ctx context.Context, uid uuid.UUID) (*repository.Wallet, error) {
			return &repository.Wallet{ID: walletID, UserID: uid, Balance: 10, Status: "active"}, nil
		},
		creditFn: func(ctx context.Context, wid uuid.UUID, amount float64, entryType, refType, description string, refID *uuid.UUID) (*repository.WalletEntry, error) {
			return nil, errors.New("credit failed")
		},
	}, nil)

	c := NewRoundUpConsumer(svc)
	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, userID.String(), mustJSON(event.RoundUpCalculated{
		TransactionID: uuid.New().String(),
		UserID:        userID.String(),
		RoundUpAmount: 0.88,
	}))
	if err == nil {
		t.Fatal("expected error for credit failure")
	}
}

func mustJSON(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}
