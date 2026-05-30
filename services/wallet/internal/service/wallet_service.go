// Copyright (c) 2026 RoundPenny. All rights reserved.

package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/wallet/internal/repository"
)

type WalletRepository interface {
	GetOrCreate(ctx context.Context, userID uuid.UUID) (*repository.Wallet, error)
	Credit(ctx context.Context, walletID uuid.UUID, amount float64, entryType, refType, description string, refID *uuid.UUID) (*repository.WalletEntry, error)
	GetEntries(ctx context.Context, userID uuid.UUID, limit, offset int) ([]repository.WalletEntry, error)
}

type WalletService struct {
	repo     WalletRepository
	producer *kafka.Producer
}

func NewWalletService(repo WalletRepository, producer *kafka.Producer) *WalletService {
	return &WalletService{repo: repo, producer: producer}
}

func (s *WalletService) GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*repository.Wallet, error) {
	return s.repo.GetOrCreate(ctx, userID)
}

func (s *WalletService) CreditRoundUp(ctx context.Context, userID uuid.UUID, amount float64, roundUpID uuid.UUID) (*repository.WalletEntry, error) {
	wallet, err := s.repo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}

	entry, err := s.repo.Credit(ctx, wallet.ID, amount, "credit", "roundup", "Round-up credit", &roundUpID)
	if err != nil {
		return nil, fmt.Errorf("credit: %w", err)
	}

	return entry, nil
}

func (s *WalletService) GetTransactions(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]repository.WalletEntry, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.GetEntries(ctx, userID, pageSize, offset)
}

func (s *WalletService) Withdraw(ctx context.Context, userID uuid.UUID, amount float64, destination map[string]any) error {
	wallet, err := s.repo.GetOrCreate(ctx, userID)
	if err != nil {
		return fmt.Errorf("get wallet: %w", err)
	}

	if wallet.Balance < amount {
		return fmt.Errorf("insufficient balance: %.2f < %.2f", wallet.Balance, amount)
	}

	_, err = s.repo.Credit(ctx, wallet.ID, -amount, "debit", "withdrawal", "Withdrawal", nil)
	return err
}
