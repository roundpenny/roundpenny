package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/investment/internal/repository"
)

type mockInvRepo struct {
	getOrCreatePortfolioFn func(ctx context.Context, userID uuid.UUID, strategy string) (*repository.Portfolio, error)
	createInvestmentFn    func(ctx context.Context, inv *repository.Investment) error
	addToPortfolioBalanceFn func(ctx context.Context, portfolioID uuid.UUID, amount float64) error
}

func (m *mockInvRepo) GetOrCreatePortfolio(ctx context.Context, userID uuid.UUID, strategy string) (*repository.Portfolio, error) {
	return m.getOrCreatePortfolioFn(ctx, userID, strategy)
}
func (m *mockInvRepo) CreateInvestment(ctx context.Context, inv *repository.Investment) error {
	return m.createInvestmentFn(ctx, inv)
}
func (m *mockInvRepo) AddToPortfolioBalance(ctx context.Context, portfolioID uuid.UUID, amount float64) error {
	return m.addToPortfolioBalanceFn(ctx, portfolioID, amount)
}

type mockInvProducer struct {
	publishFn func(ctx context.Context, msg kafka.Message) error
}

func (m *mockInvProducer) Publish(ctx context.Context, msg kafka.Message) error {
	return m.publishFn(ctx, msg)
}
func (m *mockInvProducer) Close() error { return nil }

func TestInvestRoundUp_Success(t *testing.T) {
	portfolioID := uuid.New()
	userID := uuid.New()

	var published bool
	svc := NewInvestmentService(&mockInvRepo{
		getOrCreatePortfolioFn: func(ctx context.Context, uid uuid.UUID, strategy string) (*repository.Portfolio, error) {
			return &repository.Portfolio{
				ID:       portfolioID,
				UserID:   uid,
				Strategy: strategy,
				Balance:  0,
			}, nil
		},
		createInvestmentFn: func(ctx context.Context, inv *repository.Investment) error {
			if inv.Amount != 0.792 {
				t.Fatalf("expected amount 0.792 (90%% of 0.88), got %.3f", inv.Amount)
			}
			if inv.Source != "roundup" {
				t.Fatalf("expected source roundup, got %s", inv.Source)
			}
			if inv.Status != "invested" {
				t.Fatalf("expected status invested, got %s", inv.Status)
			}
			return nil
		},
		addToPortfolioBalanceFn: func(ctx context.Context, pid uuid.UUID, amount float64) error {
			if amount != 0.792 {
				t.Fatalf("expected balance add 0.792, got %.3f", amount)
			}
			return nil
		},
	}, &mockInvProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			published = true
			if msg.Topic != event.TopicInvestmentCreated {
				t.Fatalf("expected topic %s, got %s", event.TopicInvestmentCreated, msg.Topic)
			}
			return nil
		},
	})

	err := svc.InvestRoundUp(context.Background(), userID, 0.88)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !published {
		t.Fatal("expected event to be published")
	}
}

func TestInvestRoundUp_ZeroAmount(t *testing.T) {
	userID := uuid.New()

	svc := NewInvestmentService(&mockInvRepo{
		getOrCreatePortfolioFn: func(ctx context.Context, uid uuid.UUID, strategy string) (*repository.Portfolio, error) {
			return &repository.Portfolio{
				ID:       uuid.New(),
				UserID:   uid,
				Strategy: strategy,
				Balance:  0,
			}, nil
		},
		createInvestmentFn: func(ctx context.Context, inv *repository.Investment) error {
			if inv.Amount != 0 {
				t.Fatalf("expected 0 for zero roundup, got %.3f", inv.Amount)
			}
			return nil
		},
		addToPortfolioBalanceFn: func(ctx context.Context, pid uuid.UUID, amount float64) error {
			return nil
		},
	}, &mockInvProducer{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			return nil
		},
	})

	err := svc.InvestRoundUp(context.Background(), userID, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
