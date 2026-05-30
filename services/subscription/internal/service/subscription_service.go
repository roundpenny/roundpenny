package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/subscription/internal/repository"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrPlanNotFound         = errors.New("plan not found")
	ErrAlreadyActive        = errors.New("user already has an active subscription")
	ErrInvalidTransition    = errors.New("invalid subscription status transition")
)

type SubscriptionRepository interface {
	Create(ctx context.Context, s *repository.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*repository.Subscription, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*repository.Subscription, error)
	GetActiveSubscription(ctx context.Context, userID uuid.UUID) (*repository.Subscription, error)
	Update(ctx context.Context, s *repository.Subscription) error
	List(ctx context.Context) ([]*repository.Subscription, error)
	ListPlans(ctx context.Context) ([]*repository.SubscriptionPlan, error)
	GetPlanByID(ctx context.Context, id uuid.UUID) (*repository.SubscriptionPlan, error)
	CreateBillingRecord(ctx context.Context, b *repository.BillingRecord) error
	ListBillingByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repository.BillingRecord, error)
	ListDueForRenewal(ctx context.Context, now time.Time) ([]*repository.Subscription, error)
	UpdateBillingStatus(ctx context.Context, id uuid.UUID, status string, paidAt *time.Time) error
}

type SubscriptionProducer interface {
	Publish(ctx context.Context, msg kafka.Message) error
}

type SubscriptionService struct {
	repo     SubscriptionRepository
	producer SubscriptionProducer
}

func NewSubscriptionService(repo SubscriptionRepository, producer SubscriptionProducer) *SubscriptionService {
	return &SubscriptionService{repo: repo, producer: producer}
}

type CreateSubscriptionRequest struct {
	UserID uuid.UUID
	PlanID uuid.UUID
}

type SubscriptionResponse struct {
	ID                   string         `json:"id"`
	UserID               string         `json:"user_id"`
	PlanID               string         `json:"plan_id"`
	Status               string         `json:"status"`
	CurrentPeriodStart   string         `json:"current_period_start"`
	CurrentPeriodEnd     string         `json:"current_period_end"`
	CancelledAt          *string        `json:"cancelled_at,omitempty"`
	TrialEnd             *string        `json:"trial_end,omitempty"`
	StripeSubscriptionID *string        `json:"stripe_subscription_id,omitempty"`
	Metadata             map[string]any `json:"metadata"`
	CreatedAt            string         `json:"created_at"`
	UpdatedAt            string         `json:"updated_at"`
}

type PlanResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Amount      float64  `json:"amount"`
	Currency    string   `json:"currency"`
	Interval    string   `json:"interval"`
	Features    []string `json:"features"`
	IsActive    bool     `json:"is_active"`
}

type BillingResponse struct {
	ID              string  `json:"id"`
	UserID          string  `json:"user_id"`
	SubscriptionID  *string `json:"subscription_id,omitempty"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Status          string  `json:"status"`
	PaymentMethod   string  `json:"payment_method,omitempty"`
	StripeInvoiceID *string `json:"stripe_invoice_id,omitempty"`
	PeriodStart     string  `json:"period_start"`
	PeriodEnd       string  `json:"period_end"`
	PaidAt          *string `json:"paid_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*SubscriptionResponse, error) {
	existing, err := s.repo.GetActiveSubscription(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("check active: %w", err)
	}
	if existing != nil {
		return nil, ErrAlreadyActive
	}

	plan, err := s.repo.GetPlanByID(ctx, req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("get plan: %w", err)
	}
	if plan == nil {
		return nil, ErrPlanNotFound
	}

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)
	if plan.Interval == "year" {
		periodEnd = now.AddDate(1, 0, 0)
	}

	sub := &repository.Subscription{
		UserID:             req.UserID,
		PlanID:             req.PlanID,
		Status:             "active",
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
		Metadata:           map[string]any{},
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}

	s.publishEvent(ctx, event.TopicSubscriptionCreated, event.SubscriptionCreated{
		SubscriptionID: sub.ID.String(),
		UserID:         sub.UserID.String(),
		PlanID:         sub.PlanID.String(),
		Amount:         plan.Amount,
		Currency:       plan.Currency,
	})

	return toSubscriptionResponse(sub), nil
}

func (s *SubscriptionService) CancelSubscription(ctx context.Context, id, userID uuid.UUID) (*SubscriptionResponse, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}
	if sub.UserID != userID {
		return nil, ErrSubscriptionNotFound
	}
	if sub.Status != "active" {
		return nil, ErrInvalidTransition
	}

	now := time.Now()
	sub.Status = "inactive"
	sub.CancelledAt = &now

	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, fmt.Errorf("cancel subscription: %w", err)
	}

	s.publishEvent(ctx, event.TopicSubscriptionCancelled, event.SubscriptionCancelled{
		SubscriptionID: sub.ID.String(),
		UserID:         sub.UserID.String(),
	})

	return toSubscriptionResponse(sub), nil
}

func (s *SubscriptionService) GetCurrentSubscription(ctx context.Context, userID uuid.UUID) (*SubscriptionResponse, error) {
	sub, err := s.repo.GetActiveSubscription(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get active: %w", err)
	}
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}
	return toSubscriptionResponse(sub), nil
}

func (s *SubscriptionService) ListPlans(ctx context.Context) ([]*PlanResponse, error) {
	plans, err := s.repo.ListPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}

	resp := make([]*PlanResponse, len(plans))
	for i, p := range plans {
		resp[i] = toPlanResponse(p)
	}
	return resp, nil
}

func (s *SubscriptionService) GetBillingHistory(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*BillingResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	records, err := s.repo.ListBillingByUser(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("list billing: %w", err)
	}

	resp := make([]*BillingResponse, len(records))
	for i, b := range records {
		resp[i] = toBillingResponse(b)
	}
	return resp, nil
}

func (s *SubscriptionService) ProcessRenewal(ctx context.Context, subscriptionID uuid.UUID) error {
	sub, err := s.repo.GetByID(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("get subscription: %w", err)
	}
	if sub == nil {
		return ErrSubscriptionNotFound
	}
	if sub.Status != "active" {
		return nil
	}

	plan, err := s.repo.GetPlanByID(ctx, sub.PlanID)
	if err != nil {
		return fmt.Errorf("get plan: %w", err)
	}
	if plan == nil {
		return ErrPlanNotFound
	}

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)
	if plan.Interval == "year" {
		periodEnd = now.AddDate(1, 0, 0)
	}

	billing := &repository.BillingRecord{
		UserID:         sub.UserID,
		SubscriptionID: &sub.ID,
		Amount:         plan.Amount,
		Currency:       plan.Currency,
		Status:         "pending",
		PeriodStart:    now,
		PeriodEnd:      periodEnd,
	}

	if err := s.repo.CreateBillingRecord(ctx, billing); err != nil {
		return fmt.Errorf("create billing: %w", err)
	}

	if plan.Amount > 0 {
		s.publishEvent(ctx, event.TopicPaymentFailed, event.PaymentFailed{
			SubscriptionID: sub.ID.String(),
			UserID:         sub.UserID.String(),
			Amount:         plan.Amount,
			Currency:       plan.Currency,
		})
	} else {
		paidAt := now
		if err := s.repo.UpdateBillingStatus(ctx, billing.ID, "paid", &paidAt); err != nil {
			return fmt.Errorf("update billing: %w", err)
		}
	}

	sub.CurrentPeriodStart = now
	sub.CurrentPeriodEnd = periodEnd
	if err := s.repo.Update(ctx, sub); err != nil {
		return fmt.Errorf("update period: %w", err)
	}

	s.publishEvent(ctx, event.TopicSubscriptionRenewed, event.SubscriptionRenewed{
		SubscriptionID: sub.ID.String(),
		UserID:         sub.UserID.String(),
		Amount:         plan.Amount,
		Currency:       plan.Currency,
	})

	return nil
}

func (s *SubscriptionService) publishEvent(ctx context.Context, topic string, data any) {
	if err := s.producer.Publish(ctx, kafka.Message{
		Topic:   topic,
		Payload: data,
	}); err != nil {
		slog.Warn("publish event warning", "topic", topic, "error", err)
	}
}

func toSubscriptionResponse(s *repository.Subscription) *SubscriptionResponse {
	r := &SubscriptionResponse{
		ID:                   s.ID.String(),
		UserID:               s.UserID.String(),
		PlanID:               s.PlanID.String(),
		Status:               s.Status,
		CurrentPeriodStart:   s.CurrentPeriodStart.Format(time.RFC3339),
		CurrentPeriodEnd:     s.CurrentPeriodEnd.Format(time.RFC3339),
		Metadata:             s.Metadata,
		CreatedAt:            s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            s.UpdatedAt.Format(time.RFC3339),
	}
	if s.CancelledAt != nil {
		v := s.CancelledAt.Format(time.RFC3339)
		r.CancelledAt = &v
	}
	if s.TrialEnd != nil {
		v := s.TrialEnd.Format(time.RFC3339)
		r.TrialEnd = &v
	}
	if s.StripeSubscriptionID != nil {
		r.StripeSubscriptionID = s.StripeSubscriptionID
	}
	return r
}

func toPlanResponse(p *repository.SubscriptionPlan) *PlanResponse {
	return &PlanResponse{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		Amount:      p.Amount,
		Currency:    p.Currency,
		Interval:    p.Interval,
		Features:    p.Features,
		IsActive:    p.IsActive,
	}
}

func toBillingResponse(b *repository.BillingRecord) *BillingResponse {
	r := &BillingResponse{
		ID:            b.ID.String(),
		UserID:        b.UserID.String(),
		Amount:        b.Amount,
		Currency:      b.Currency,
		Status:        b.Status,
		PaymentMethod: b.PaymentMethod,
		PeriodStart:   b.PeriodStart.Format(time.RFC3339),
		PeriodEnd:     b.PeriodEnd.Format(time.RFC3339),
		CreatedAt:     b.CreatedAt.Format(time.RFC3339),
	}
	if b.SubscriptionID != nil {
		v := b.SubscriptionID.String()
		r.SubscriptionID = &v
	}
	if b.StripeInvoiceID != nil {
		r.StripeInvoiceID = b.StripeInvoiceID
	}
	if b.PaidAt != nil {
		v := b.PaidAt.Format(time.RFC3339)
		r.PaidAt = &v
	}
	return r
}
