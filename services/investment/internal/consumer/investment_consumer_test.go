package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/investment/internal/repository"
	"github.com/roundup-platform/services/investment/internal/service"
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

type mockInvProd struct {
	publishFn func(ctx context.Context, msg kafka.Message) error
}

func (m *mockInvProd) Publish(ctx context.Context, msg kafka.Message) error {
	return m.publishFn(ctx, msg)
}
func (m *mockInvProd) Close() error { return nil }

func TestInvestmentConsumer_HandleRoundUp_Success(t *testing.T) {
	userID := uuid.New()
	var invested bool

	svc := service.NewInvestmentService(&mockInvRepo{
		getOrCreatePortfolioFn: func(ctx context.Context, uid uuid.UUID, strategy string) (*repository.Portfolio, error) {
			return &repository.Portfolio{ID: uuid.New(), UserID: uid, Strategy: strategy, Balance: 0}, nil
		},
		createInvestmentFn: func(ctx context.Context, inv *repository.Investment) error {
			invested = true
			return nil
		},
		addToPortfolioBalanceFn: func(ctx context.Context, pid uuid.UUID, amount float64) error {
			return nil
		},
	}, &mockInvProd{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			return nil
		},
	})

	c := NewInvestmentConsumer(svc)
	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, userID.String(), mustJSON(event.RoundUpCalculated{
		TransactionID: uuid.New().String(),
		UserID:        userID.String(),
		RoundUpAmount: 0.88,
		Currency:      "USD",
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !invested {
		t.Fatal("expected investment to be created")
	}
}

func TestInvestmentConsumer_HandleRoundUp_InvalidJSON(t *testing.T) {
	svc := service.NewInvestmentService(&mockInvRepo{}, &mockInvProd{})
	c := NewInvestmentConsumer(svc)

	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, uuid.New().String(), []byte("not-json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestInvestmentConsumer_HandleRoundUp_InvalidUserID(t *testing.T) {
	svc := service.NewInvestmentService(&mockInvRepo{}, &mockInvProd{})
	c := NewInvestmentConsumer(svc)

	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, uuid.New().String(), mustJSON(event.RoundUpCalculated{
		TransactionID: uuid.New().String(),
		UserID:        "not-a-uuid",
		RoundUpAmount: 0.88,
	}))
	if err == nil {
		t.Fatal("expected error for invalid user ID")
	}
}

func TestInvestmentConsumer_HandleRoundUp_PortfolioError(t *testing.T) {
	userID := uuid.New()

	svc := service.NewInvestmentService(&mockInvRepo{
		getOrCreatePortfolioFn: func(ctx context.Context, uid uuid.UUID, strategy string) (*repository.Portfolio, error) {
			return nil, errors.New("db connection failed")
		},
	}, &mockInvProd{})

	c := NewInvestmentConsumer(svc)
	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, userID.String(), mustJSON(event.RoundUpCalculated{
		TransactionID: uuid.New().String(),
		UserID:        userID.String(),
		RoundUpAmount: 0.88,
	}))
	if err != nil {
		t.Fatalf("expected no error (consumer swallows), got %v", err)
	}
}

func TestInvestmentConsumer_HandleRoundUp_CreateInvestmentError(t *testing.T) {
	userID := uuid.New()

	svc := service.NewInvestmentService(&mockInvRepo{
		getOrCreatePortfolioFn: func(ctx context.Context, uid uuid.UUID, strategy string) (*repository.Portfolio, error) {
			return &repository.Portfolio{ID: uuid.New(), UserID: uid, Strategy: strategy, Balance: 0}, nil
		},
		createInvestmentFn: func(ctx context.Context, inv *repository.Investment) error {
			return errors.New("duplicate investment")
		},
		addToPortfolioBalanceFn: func(ctx context.Context, pid uuid.UUID, amount float64) error {
			return nil
		},
	}, &mockInvProd{
		publishFn: func(ctx context.Context, msg kafka.Message) error {
			return nil
		},
	})

	c := NewInvestmentConsumer(svc)
	err := c.HandleRoundUp(context.Background(), event.TopicRoundUpCalculated, userID.String(), mustJSON(event.RoundUpCalculated{
		TransactionID: uuid.New().String(),
		UserID:        userID.String(),
		RoundUpAmount: 0.88,
	}))
	if err != nil {
		t.Fatalf("expected no error (consumer swallows), got %v", err)
	}
}

func mustJSON(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}
