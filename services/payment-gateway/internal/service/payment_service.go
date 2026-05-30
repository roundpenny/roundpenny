// Copyright (c) 2026 RoundPenny. All rights reserved.

package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	stripepkg "github.com/roundup-platform/pkg/stripe"
	"github.com/roundup-platform/services/payment-gateway/internal/repository"
)

var (
	ErrPaymentNotFound    = errors.New("payment not found")
	ErrInvalidAmount      = errors.New("invalid payment amount")
	ErrPaymentFailed      = errors.New("payment failed")
	ErrInvalidStatus      = errors.New("invalid payment status transition")
)

type PaymentRepository interface {
	Create(ctx context.Context, p *repository.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*repository.Payment, error)
	GetByStripeIntentID(ctx context.Context, stripeIntentID string) (*repository.Payment, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repository.Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, transactionID *uuid.UUID, errorMessage string) error
	UpdateStripePaymentMethod(ctx context.Context, id uuid.UUID, stripePaymentMethodID string) error
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*repository.Payment, error)
}

type CreatePaymentRequest struct {
	UserID                uuid.UUID
	Amount                float64
	Currency              string
	PaymentMethod         string
	StripePaymentIntentID string
	Description           string
	Metadata              map[string]any
}

type PaymentResponse struct {
	ID                    uuid.UUID      `json:"id"`
	UserID                uuid.UUID      `json:"user_id"`
	TransactionID         *uuid.UUID     `json:"transaction_id,omitempty"`
	Amount                float64        `json:"amount"`
	Currency              string         `json:"currency"`
	Status                string         `json:"status"`
	PaymentMethod         string         `json:"payment_method,omitempty"`
	StripePaymentIntentID string         `json:"stripe_payment_intent_id,omitempty"`
	ClientSecret          string         `json:"client_secret,omitempty"`
	Description           string         `json:"description,omitempty"`
	Metadata              map[string]any `json:"metadata,omitempty"`
	ErrorMessage          string         `json:"error_message,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

type PaymentService struct {
	repo          PaymentRepository
	stripeClient  *stripepkg.StripeClient
}

func NewPaymentService(repo PaymentRepository, stripeClient *stripepkg.StripeClient) *PaymentService {
	return &PaymentService{repo: repo, stripeClient: stripeClient}
}

func (s *PaymentService) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResponse, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}

	piID := req.StripePaymentIntentID
	clientSecret := ""

	if s.stripeClient != nil {
		pi, err := s.stripeClient.CreatePaymentIntent(stripepkg.CreatePaymentIntentParams{
			Amount:        int64(req.Amount * 100),
			Currency:      req.Currency,
			Description:   req.Description,
			PaymentMethod: req.PaymentMethod,
			Metadata: map[string]string{
				"user_id": req.UserID.String(),
			},
		})
		if err != nil {
			return nil, err
		}
		piID = pi.ID
		clientSecret = pi.ClientSecret
	}

	p := &repository.Payment{
		UserID:                req.UserID,
		Amount:                req.Amount,
		Currency:              req.Currency,
		Status:                "pending",
		PaymentMethod:         req.PaymentMethod,
		StripePaymentIntentID: piID,
		Description:           req.Description,
		Metadata:              req.Metadata,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	resp := toResponse(p)
	resp.ClientSecret = clientSecret
	resp.StripePaymentIntentID = piID
	return resp, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, id uuid.UUID) (*PaymentResponse, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPaymentNotFound
	}
	return toResponse(p), nil
}

func (s *PaymentService) GetPaymentByStripeIntent(ctx context.Context, stripeIntentID string) (*PaymentResponse, error) {
	p, err := s.repo.GetByStripeIntentID(ctx, stripeIntentID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPaymentNotFound
	}
	return toResponse(p), nil
}

func (s *PaymentService) ListUserPayments(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*PaymentResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	payments, err := s.repo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*PaymentResponse, len(payments))
	for i, p := range payments {
		responses[i] = toResponse(p)
	}
	return responses, nil
}

func (s *PaymentService) ConfirmPayment(ctx context.Context, id uuid.UUID, stripePaymentMethodID string) (*PaymentResponse, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPaymentNotFound
	}

	if p.Status != "pending" {
		return nil, ErrInvalidStatus
	}

	txID := uuid.New()
	if err := s.repo.UpdateStatus(ctx, id, "completed", &txID, ""); err != nil {
		return nil, err
	}

	if s.stripeClient != nil {
		s.stripeClient.ConfirmPaymentIntent(p.StripePaymentIntentID, stripePaymentMethodID)
	}

	if err := s.repo.UpdateStripePaymentMethod(ctx, id, stripePaymentMethodID); err != nil {
		return nil, err
	}

	return s.GetPayment(ctx, id)
}

func (s *PaymentService) FailPayment(ctx context.Context, id uuid.UUID, errorMessage string) (*PaymentResponse, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPaymentNotFound
	}

	if p.Status != "pending" {
		return nil, ErrInvalidStatus
	}

	if err := s.repo.UpdateStatus(ctx, id, "failed", nil, errorMessage); err != nil {
		return nil, err
	}

	return s.GetPayment(ctx, id)
}

func (s *PaymentService) SucceedPayment(ctx context.Context, id uuid.UUID, transactionID uuid.UUID) (*PaymentResponse, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPaymentNotFound
	}

	if p.Status != "pending" {
		return nil, ErrInvalidStatus
	}

	txID := &transactionID
	if err := s.repo.UpdateStatus(ctx, id, "completed", txID, ""); err != nil {
		return nil, err
	}

	return s.GetPayment(ctx, id)
}

func toResponse(p *repository.Payment) *PaymentResponse {
	return &PaymentResponse{
		ID:                    p.ID,
		UserID:                p.UserID,
		TransactionID:         p.TransactionID,
		Amount:                p.Amount,
		Currency:              p.Currency,
		Status:                p.Status,
		PaymentMethod:         p.PaymentMethod,
		StripePaymentIntentID: p.StripePaymentIntentID,
		Description:           p.Description,
		Metadata:              p.Metadata,
		ErrorMessage:          p.ErrorMessage,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}
}
