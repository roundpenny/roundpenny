package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/services/ledger/internal/repository"
)

type mockLedgerRepo struct {
	createDoubleEntryFn func(ctx context.Context, description string, txID, roundUpID *uuid.UUID, lines []repository.JournalLine) error
	getEntryFn          func(ctx context.Context, id uuid.UUID) (*repository.JournalEntry, error)
	listByAccountFn     func(ctx context.Context, accountCode string, limit, offset int) ([]repository.JournalLine, error)
	listByUserFn        func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]repository.JournalLine, error)
}

func (m *mockLedgerRepo) CreateDoubleEntry(ctx context.Context, description string, txID, roundUpID *uuid.UUID, lines []repository.JournalLine) error {
	return m.createDoubleEntryFn(ctx, description, txID, roundUpID, lines)
}

func (m *mockLedgerRepo) GetEntry(ctx context.Context, id uuid.UUID) (*repository.JournalEntry, error) {
	return m.getEntryFn(ctx, id)
}

func (m *mockLedgerRepo) ListByAccount(ctx context.Context, accountCode string, limit, offset int) ([]repository.JournalLine, error) {
	return m.listByAccountFn(ctx, accountCode, limit, offset)
}

func (m *mockLedgerRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]repository.JournalLine, error) {
	return m.listByUserFn(ctx, userID, limit, offset)
}

func TestCreateEntry_Success(t *testing.T) {
	txID := uuid.New()
	userID := uuid.New()

	var called bool
	svc := NewLedgerService(&mockLedgerRepo{
		createDoubleEntryFn: func(ctx context.Context, description string, txID_, roundUpID *uuid.UUID, lines []repository.JournalLine) error {
			called = true
			if description != "Round-up credit" {
				t.Fatalf("expected description 'Round-up credit', got %q", description)
			}
			if len(lines) != 3 {
				t.Fatalf("expected 3 lines, got %d", len(lines))
			}
			return nil
		},
	}, nil)

	err := svc.RecordRoundUp(context.Background(), event.RoundUpCalculated{
		TransactionID: txID.String(),
		UserID:        userID.String(),
		RoundUpAmount: 0.42,
		Currency:      "USD",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected CreateDoubleEntry to be called")
	}
}

func TestCreateEntry_MissingRequiredFields(t *testing.T) {
	var called bool
	svc := NewLedgerService(&mockLedgerRepo{
		createDoubleEntryFn: func(ctx context.Context, description string, txID, roundUpID *uuid.UUID, lines []repository.JournalLine) error {
			called = true
			return nil
		},
	}, nil)

	err := svc.RecordRoundUp(context.Background(), event.RoundUpCalculated{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected CreateDoubleEntry to be called")
	}
}

func TestGetEntry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		entryID := uuid.New()
		svc := NewLedgerService(&mockLedgerRepo{
			getEntryFn: func(ctx context.Context, id uuid.UUID) (*repository.JournalEntry, error) {
				return &repository.JournalEntry{
					ID:          id,
					Description: "Test entry",
				}, nil
			},
		}, nil)

		entry, err := svc.GetEntry(context.Background(), entryID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if entry.ID != entryID {
			t.Fatalf("expected ID %v, got %v", entryID, entry.ID)
		}
		if entry.Description != "Test entry" {
			t.Fatalf("expected description 'Test entry', got %q", entry.Description)
		}
	})

	t.Run("not found", func(t *testing.T) {
		entryID := uuid.New()
		svc := NewLedgerService(&mockLedgerRepo{
			getEntryFn: func(ctx context.Context, id uuid.UUID) (*repository.JournalEntry, error) {
				return nil, nil
			},
		}, nil)

		entry, err := svc.GetEntry(context.Background(), entryID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if entry != nil {
			t.Fatal("expected nil entry for not found")
		}
	})
}

func TestListByAccount_Pagination(t *testing.T) {
	svc := NewLedgerService(&mockLedgerRepo{
		listByAccountFn: func(ctx context.Context, accountCode string, limit, offset int) ([]repository.JournalLine, error) {
			if accountCode != "1000" {
				t.Fatalf("expected account code '1000', got %q", accountCode)
			}
			if limit != 10 {
				t.Fatalf("expected limit 10, got %d", limit)
			}
			if offset != 10 {
				t.Fatalf("expected offset 10 for page 2, got %d", offset)
			}
			return []repository.JournalLine{
				{AccountCode: "1000", DebitAmount: 100, CreditAmount: 0, Currency: "USD"},
			}, nil
		},
	}, nil)

	lines, err := svc.ListByAccount(context.Background(), "1000", 2, 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
}

func TestListByUser_Pagination(t *testing.T) {
	userID := uuid.New()

	svc := NewLedgerService(&mockLedgerRepo{
		listByUserFn: func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]repository.JournalLine, error) {
			if uid != userID {
				t.Fatalf("expected userID %v, got %v", userID, uid)
			}
			if limit != 20 {
				t.Fatalf("expected default limit 20, got %d", limit)
			}
			if offset != 0 {
				t.Fatalf("expected offset 0 for page 1, got %d", offset)
			}
			return []repository.JournalLine{
				{AccountCode: "1100", DebitAmount: 0.42, CreditAmount: 0, Currency: "USD", UserID: &uid},
			}, nil
		},
	}, nil)

	lines, err := svc.ListByUser(context.Background(), userID, 0, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
}
