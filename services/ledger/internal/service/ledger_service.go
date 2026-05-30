// Copyright (c) 2026 RoundPenny. All rights reserved.

package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/ledger/internal/repository"
)

type LedgerRepository interface {
	CreateDoubleEntry(ctx context.Context, description string, txID, roundUpID *uuid.UUID, lines []repository.JournalLine) error
	GetEntry(ctx context.Context, id uuid.UUID) (*repository.JournalEntry, error)
	ListByAccount(ctx context.Context, accountCode string, limit, offset int) ([]repository.JournalLine, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]repository.JournalLine, error)
}

type LedgerProducer interface {
	Publish(ctx context.Context, msg kafka.Message) error
}

type LedgerService struct {
	repo     LedgerRepository
	producer LedgerProducer
}

func NewLedgerService(repo LedgerRepository, producer LedgerProducer) *LedgerService {
	return &LedgerService{repo: repo, producer: producer}
}

func (s *LedgerService) RecordRoundUp(ctx context.Context, evt event.RoundUpCalculated) error {
	txID, _ := uuid.Parse(evt.TransactionID)
	userID, _ := uuid.Parse(evt.UserID)

	lines := []repository.JournalLine{
		{AccountCode: "1100", DebitAmount: evt.RoundUpAmount, CreditAmount: 0, Currency: evt.Currency, UserID: &userID},
		{AccountCode: "1200", CreditAmount: evt.RoundUpAmount * 0.90, DebitAmount: 0, Currency: evt.Currency, UserID: &userID},
		{AccountCode: "2000", CreditAmount: evt.RoundUpAmount * 0.10, DebitAmount: 0, Currency: evt.Currency},
	}

	return s.repo.CreateDoubleEntry(ctx, "Round-up credit", &txID, &txID, lines)
}

func (s *LedgerService) RecordFee(ctx context.Context, evt event.FeeCharged) error {
	feeTxID, _ := uuid.Parse(evt.TransactionID)
	feeUserID, _ := uuid.Parse(evt.UserID)

	lines := []repository.JournalLine{
		{AccountCode: "1000", DebitAmount: evt.Amount, CreditAmount: 0, Currency: "USD"},
		{AccountCode: "2000", CreditAmount: evt.Amount, DebitAmount: 0, Currency: "USD", UserID: &feeUserID},
	}

	return s.repo.CreateDoubleEntry(ctx, "Fee charge", &feeTxID, nil, lines)
}

func (s *LedgerService) RecordInvestment(ctx context.Context, evt event.InvestmentCreated) error {
	slog.Info("ledger: investment recorded", "user", evt.UserID, "amount", evt.Amount)
	return nil
}

func (s *LedgerService) GetEntry(ctx context.Context, id uuid.UUID) (*repository.JournalEntry, error) {
	return s.repo.GetEntry(ctx, id)
}

func (s *LedgerService) ListByAccount(ctx context.Context, accountCode string, page, pageSize int) ([]repository.JournalLine, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListByAccount(ctx, accountCode, pageSize, offset)
}

func (s *LedgerService) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]repository.JournalLine, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListByUser(ctx, userID, pageSize, offset)
}
