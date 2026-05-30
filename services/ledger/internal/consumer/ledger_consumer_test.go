package consumer

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/services/ledger/internal/repository"
	"github.com/roundup-platform/services/ledger/internal/service"
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

func TestLedgerConsumer_HandleEvent_RoundUpCalculated(t *testing.T) {
	var called bool
	svc := service.NewLedgerService(&mockLedgerRepo{
		createDoubleEntryFn: func(ctx context.Context, desc string, txID, roundUpID *uuid.UUID, lines []repository.JournalLine) error {
			called = true
			if len(lines) != 3 {
				t.Fatalf("expected 3 journal lines, got %d", len(lines))
			}
			return nil
		},
	}, nil)

	c := NewLedgerConsumer(svc)
	err := c.HandleEvent(context.Background(), event.TopicRoundUpCalculated, uuid.New().String(), mustJSON(event.RoundUpCalculated{
		TransactionID: uuid.New().String(),
		UserID:        uuid.New().String(),
		RoundUpAmount: 0.42,
		Currency:      "USD",
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected CreateDoubleEntry to be called")
	}
}

func TestLedgerConsumer_HandleEvent_FeeCharged(t *testing.T) {
	var called bool
	svc := service.NewLedgerService(&mockLedgerRepo{
		createDoubleEntryFn: func(ctx context.Context, desc string, txID, roundUpID *uuid.UUID, lines []repository.JournalLine) error {
			called = true
			if desc != "Fee charge" {
				t.Fatalf("expected 'Fee charge', got %q", desc)
			}
			if len(lines) != 2 {
				t.Fatalf("expected 2 journal lines, got %d", len(lines))
			}
			return nil
		},
	}, nil)

	c := NewLedgerConsumer(svc)
	err := c.HandleEvent(context.Background(), event.TopicFeeCharged, uuid.New().String(), mustJSON(event.FeeCharged{
		TransactionID: uuid.New().String(),
		UserID:        uuid.New().String(),
		Amount:        0.10,
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected CreateDoubleEntry to be called for fee")
	}
}

func TestLedgerConsumer_HandleEvent_InvestmentCreated(t *testing.T) {
	svc := service.NewLedgerService(&mockLedgerRepo{}, nil)
	c := NewLedgerConsumer(svc)
	err := c.HandleEvent(context.Background(), event.TopicInvestmentCreated, uuid.New().String(), mustJSON(event.InvestmentCreated{
		UserID: uuid.New().String(),
		Amount: 0.50,
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestLedgerConsumer_HandleEvent_UnknownTopic(t *testing.T) {
	svc := service.NewLedgerService(&mockLedgerRepo{}, nil)
	c := NewLedgerConsumer(svc)
	err := c.HandleEvent(context.Background(), "unknown.topic", uuid.New().String(), []byte(`{}`))
	if err != nil {
		t.Fatalf("expected no error for unknown topic, got %v", err)
	}
}

func TestLedgerConsumer_HandleEvent_InvalidJSON(t *testing.T) {
	svc := service.NewLedgerService(&mockLedgerRepo{}, nil)
	c := NewLedgerConsumer(svc)

	err := c.HandleEvent(context.Background(), event.TopicRoundUpCalculated, uuid.New().String(), []byte("not-json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLedgerConsumer_HandleEvent_EmptyPayload(t *testing.T) {
	svc := service.NewLedgerService(&mockLedgerRepo{}, nil)
	c := NewLedgerConsumer(svc)

	err := c.HandleEvent(context.Background(), event.TopicRoundUpCalculated, uuid.New().String(), []byte{})
	if err == nil {
		t.Fatal("expected error for empty payload")
	}
}

func TestLedgerConsumer_ConcurrentMessages(t *testing.T) {
	var mu sync.Mutex
	var count int

	svc := service.NewLedgerService(&mockLedgerRepo{
		createDoubleEntryFn: func(ctx context.Context, desc string, txID, roundUpID *uuid.UUID, lines []repository.JournalLine) error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		},
	}, nil)

	c := NewLedgerConsumer(svc)
	n := 10
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_ = c.HandleEvent(context.Background(), event.TopicRoundUpCalculated, uuid.New().String(), mustJSON(event.RoundUpCalculated{
				TransactionID: uuid.New().String(),
				UserID:        uuid.New().String(),
				RoundUpAmount: 0.42,
				Currency:      "USD",
			}))
		}()
	}

	wg.Wait()
	if count != n {
		t.Fatalf("expected %d calls, got %d", n, count)
	}
}

func mustJSON(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}
