// Copyright (c) 2026 RoundPenny. All rights reserved.

package models

import (
	"time"
	"github.com/google/uuid"
)

type Subscription struct {
	ID                  uuid.UUID       `json:"id"`
	UserID              uuid.UUID       `json:"user_id"`
	PlanID              uuid.UUID       `json:"plan_id"`
	Status              string          `json:"status"`
	CurrentPeriodStart  time.Time       `json:"current_period_start"`
	CurrentPeriodEnd    time.Time       `json:"current_period_end"`
	CancelledAt         *time.Time      `json:"cancelled_at,omitempty"`
	TrialEnd            *time.Time      `json:"trial_end,omitempty"`
	StripeSubscriptionID *string        `json:"stripe_subscription_id,omitempty"`
	Metadata            map[string]any  `json:"metadata"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type SubscriptionPlan struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Amount      float64                `json:"amount"`
	Currency    string                 `json:"currency"`
	Interval    string                 `json:"interval"`
	Features    []string               `json:"features"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
}

type BillingRecord struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	SubscriptionID   *uuid.UUID `json:"subscription_id,omitempty"`
	Amount           float64    `json:"amount"`
	Currency         string     `json:"currency"`
	Status           string     `json:"status"`
	PaymentMethod    string     `json:"payment_method,omitempty"`
	StripeInvoiceID  *string    `json:"stripe_invoice_id,omitempty"`
	PeriodStart      time.Time  `json:"period_start"`
	PeriodEnd        time.Time  `json:"period_end"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}
