// Copyright (c) 2026 RoundPenny. All rights reserved.

package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/payment-gateway/internal/repository"
)

type mockPaymentRepo struct {
	createFn              func(ctx context.Context, p *repository.Payment) error
	getByIDFn             func(ctx context.Context, id uuid.UUID) (*repository.Payment, error)
	getByStripeIntentFn   func(ctx context.Context, id string) (*repository.Payment, error)
	getByUserIDFn         func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repository.Payment, error)
	updateStatusFn        func(ctx context.Context, id uuid.UUID, status string, transactionID *uuid.UUID, errorMessage string) error
	updateStripeMethodFn  func(ctx context.Context, id uuid.UUID, stripePaymentMethodID string) error
	listByStatusFn        func(ctx context.Context, status string, limit, offset int) ([]*repository.Payment, error)
}

func (m *mockPaymentRepo) Create(ctx context.Context, p *repository.Payment) error {
	return m.createFn(ctx, p)
}
func (m *mockPaymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*repository.Payment, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockPaymentRepo) GetByStripeIntentID(ctx context.Context, id string) (*repository.Payment, error) {
	return m.getByStripeIntentFn(ctx, id)
}
func (m *mockPaymentRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repository.Payment, error) {
	return m.getByUserIDFn(ctx, userID, limit, offset)
}
func (m *mockPaymentRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, transactionID *uuid.UUID, errorMessage string) error {
	return m.updateStatusFn(ctx, id, status, transactionID, errorMessage)
}
func (m *mockPaymentRepo) UpdateStripePaymentMethod(ctx context.Context, id uuid.UUID, stripePaymentMethodID string) error {
	return m.updateStripeMethodFn(ctx, id, stripePaymentMethodID)
}
func (m *mockPaymentRepo) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*repository.Payment, error) {
	return m.listByStatusFn(ctx, status, limit, offset)
}

func TestCreatePayment_Success(t *testing.T) {
	userID := uuid.New()
	paymentID := uuid.New()
	now := time.Now()

	mock := &mockPaymentRepo{
		createFn: func(ctx context.Context, p *repository.Payment) error {
			p.ID = paymentID
			p.CreatedAt = now
			p.UpdatedAt = now
			return nil
		},
	}

	svc := 	NewPaymentService(mock, nil)
	resp, err := svc.CreatePayment(context.Background(), CreatePaymentRequest{
		UserID:                userID,
		Amount:                99.99,
		Currency:              "USD",
		PaymentMethod:         "card",
		StripePaymentIntentID: "pi_test_123",
		Description:           "test payment",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Amount != 99.99 {
		t.Errorf("expected 99.99, got %f", resp.Amount)
	}
	if resp.Status != "pending" {
		t.Errorf("expected pending, got %s", resp.Status)
	}
	if resp.StripePaymentIntentID != "pi_test_123" {
		t.Errorf("expected pi_test_123, got %s", resp.StripePaymentIntentID)
	}
}

func TestCreatePayment_InvalidAmount(t *testing.T) {
	svc := NewPaymentService(&mockPaymentRepo{}, nil)
	_, err := svc.CreatePayment(context.Background(), CreatePaymentRequest{
		UserID: uuid.New(),
		Amount: 0,
	})
	if err != ErrInvalidAmount {
		t.Errorf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestGetPayment_NotFound(t *testing.T) {
	mock := &mockPaymentRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Payment, error) {
			return nil, nil
		},
	}

	svc := 	NewPaymentService(mock, nil)
	_, err := svc.GetPayment(context.Background(), uuid.New())
	if err != ErrPaymentNotFound {
		t.Errorf("expected ErrPaymentNotFound, got %v", err)
	}
}

func TestGetPayment_Success(t *testing.T) {
	paymentID := uuid.New()
	userID := uuid.New()

	mock := &mockPaymentRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Payment, error) {
			return &repository.Payment{
				ID:     paymentID,
				UserID: userID,
				Amount: 50.00,
				Status: "completed",
			}, nil
		},
	}

	svc := 	NewPaymentService(mock, nil)
	resp, err := svc.GetPayment(context.Background(), paymentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != paymentID {
		t.Errorf("expected %v, got %v", paymentID, resp.ID)
	}
	if resp.Status != "completed" {
		t.Errorf("expected completed, got %s", resp.Status)
	}
}

func TestSucceedPayment_InvalidTransition(t *testing.T) {
	paymentID := uuid.New()

	mock := &mockPaymentRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Payment, error) {
			return &repository.Payment{
				ID:     paymentID,
				Status: "completed",
			}, nil
		},
	}

	svc := 	NewPaymentService(mock, nil)
	_, err := svc.SucceedPayment(context.Background(), paymentID, uuid.New())
	if err != ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestSucceedPayment_Success(t *testing.T) {
	paymentID := uuid.New()
	txID := uuid.New()
	currentStatus := "pending"

	mock := &mockPaymentRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Payment, error) {
			return &repository.Payment{
				ID:     paymentID,
				Status: currentStatus,
			}, nil
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status string, transactionID *uuid.UUID, errorMessage string) error {
			currentStatus = status
			return nil
		},
	}

	svc := 	NewPaymentService(mock, nil)
	resp, err := svc.SucceedPayment(context.Background(), paymentID, txID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "completed" {
		t.Errorf("expected completed, got %s", resp.Status)
	}
}

func TestFailPayment_Success(t *testing.T) {
	paymentID := uuid.New()
	currentStatus := "pending"

	mock := &mockPaymentRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Payment, error) {
			return &repository.Payment{
				ID:     paymentID,
				Status: currentStatus,
			}, nil
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status string, transactionID *uuid.UUID, errorMessage string) error {
			currentStatus = status
			return nil
		},
	}

	svc := 	NewPaymentService(mock, nil)
	resp, err := svc.FailPayment(context.Background(), paymentID, "card_declined")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "failed" {
		t.Errorf("expected failed, got %s", resp.Status)
	}
}

func TestListUserPayments_DefaultPagination(t *testing.T) {
	userID := uuid.New()

	mock := &mockPaymentRepo{
		getByUserIDFn: func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*repository.Payment, error) {
			if limit != 20 || offset != 0 {
				t.Errorf("expected limit=20, offset=0, got limit=%d offset=%d", limit, offset)
			}
			return []*repository.Payment{}, nil
		},
	}

	svc := 	NewPaymentService(mock, nil)
	resp, err := svc.ListUserPayments(context.Background(), userID, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list")
	}
}

func TestGetPaymentByStripeIntent_Success(t *testing.T) {
	stripeIntentID := "pi_test_456"
	paymentID := uuid.New()

	mock := &mockPaymentRepo{
		getByStripeIntentFn: func(ctx context.Context, id string) (*repository.Payment, error) {
			return &repository.Payment{
				ID:     paymentID,
				Status: "pending",
			}, nil
		},
	}

	svc := 	NewPaymentService(mock, nil)
	resp, err := svc.GetPaymentByStripeIntent(context.Background(), stripeIntentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != paymentID {
		t.Errorf("expected %v, got %v", paymentID, resp.ID)
	}
}
