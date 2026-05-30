// Copyright (c) 2026 RoundPenny. All rights reserved.

package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/wallet/internal/repository"
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

func TestGetOrCreateWallet(t *testing.T) {
	walletID := uuid.New()
	userID := uuid.New()

	svc := NewWalletService(&mockWalletRepo{
		getOrCreateFn: func(ctx context.Context, uid uuid.UUID) (*repository.Wallet, error) {
			return &repository.Wallet{
				ID:       walletID,
				UserID:   uid,
				Balance:  0,
				Currency: "USD",
				Status:   "active",
			}, nil
		},
	}, nil)

	w, err := svc.GetOrCreateWallet(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if w.ID != walletID {
		t.Fatalf("expected wallet ID %v, got %v", walletID, w.ID)
	}
}

func TestCreditRoundUp(t *testing.T) {
	walletID := uuid.New()
	userID := uuid.New()
	roundUpID := uuid.New()

	svc := NewWalletService(&mockWalletRepo{
		getOrCreateFn: func(ctx context.Context, uid uuid.UUID) (*repository.Wallet, error) {
			return &repository.Wallet{
				ID:       walletID,
				UserID:   uid,
				Balance:  0,
				Currency: "USD",
				Status:   "active",
			}, nil
		},
		creditFn: func(ctx context.Context, wid uuid.UUID, amount float64, entryType, refType, description string, refID *uuid.UUID) (*repository.WalletEntry, error) {
			return &repository.WalletEntry{
				WalletID:      wid,
				UserID:        userID,
				Amount:        amount,
				BalanceBefore: 0,
				BalanceAfter:  amount,
				EntryType:     "credit",
				ReferenceType: "roundup",
				ReferenceID:   refID,
				Description:   description,
				CreatedAt:     time.Now(),
			}, nil
		},
	}, nil)

	entry, err := svc.CreditRoundUp(context.Background(), userID, 0.88, roundUpID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if entry.Amount != 0.88 {
		t.Fatalf("expected amount 0.88, got %.2f", entry.Amount)
	}
	if entry.BalanceAfter != 0.88 {
		t.Fatalf("expected balance after 0.88, got %.2f", entry.BalanceAfter)
	}
}

func TestWithdraw_Success(t *testing.T) {
	walletID := uuid.New()
	userID := uuid.New()

	svc := NewWalletService(&mockWalletRepo{
		getOrCreateFn: func(ctx context.Context, uid uuid.UUID) (*repository.Wallet, error) {
			return &repository.Wallet{
				ID:       walletID,
				UserID:   uid,
				Balance:  100,
				Currency: "USD",
				Status:   "active",
			}, nil
		},
		creditFn: func(ctx context.Context, wid uuid.UUID, amount float64, entryType, refType, description string, refID *uuid.UUID) (*repository.WalletEntry, error) {
			return &repository.WalletEntry{
				Amount:        amount,
				BalanceBefore: 100,
				BalanceAfter:  100 + amount,
				EntryType:     "debit",
			}, nil
		},
	}, nil)

	err := svc.Withdraw(context.Background(), userID, 30, map[string]any{"bank": "test"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestWithdraw_InsufficientBalance(t *testing.T) {
	walletID := uuid.New()
	userID := uuid.New()

	svc := NewWalletService(&mockWalletRepo{
		getOrCreateFn: func(ctx context.Context, uid uuid.UUID) (*repository.Wallet, error) {
			return &repository.Wallet{
				ID:       walletID,
				UserID:   uid,
				Balance:  10,
				Currency: "USD",
				Status:   "active",
			}, nil
		},
	}, nil)

	err := svc.Withdraw(context.Background(), userID, 30, map[string]any{})
	if err == nil {
		t.Fatal("expected error for insufficient balance")
	}
}

func TestWithdraw_ExactBalance(t *testing.T) {
	walletID := uuid.New()
	userID := uuid.New()

	svc := NewWalletService(&mockWalletRepo{
		getOrCreateFn: func(ctx context.Context, uid uuid.UUID) (*repository.Wallet, error) {
			return &repository.Wallet{
				ID:       walletID,
				UserID:   uid,
				Balance:  30,
				Currency: "USD",
				Status:   "active",
			}, nil
		},
		creditFn: func(ctx context.Context, wid uuid.UUID, amount float64, entryType, refType, description string, refID *uuid.UUID) (*repository.WalletEntry, error) {
			return &repository.WalletEntry{
				Amount: -30,
			}, nil
		},
	}, nil)

	err := svc.Withdraw(context.Background(), userID, 30, map[string]any{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGetTransactions(t *testing.T) {
	userID := uuid.New()

	svc := NewWalletService(&mockWalletRepo{
		getEntriesFn: func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]repository.WalletEntry, error) {
			return []repository.WalletEntry{
				{Amount: 0.88, EntryType: "credit", Description: "Round-up credit"},
			}, nil
		},
	}, nil)

	entries, err := svc.GetTransactions(context.Background(), userID, 1, 20)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Amount != 0.88 {
		t.Fatalf("expected amount 0.88, got %.2f", entries[0].Amount)
	}
}
