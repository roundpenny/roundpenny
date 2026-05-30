package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/roundup-platform/pkg/db"
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
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Interval    string    `json:"interval"`
	Features    []string  `json:"features"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type BillingRecord struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	SubscriptionID  *uuid.UUID `json:"subscription_id,omitempty"`
	Amount          float64    `json:"amount"`
	Currency        string     `json:"currency"`
	Status          string     `json:"status"`
	PaymentMethod   string     `json:"payment_method,omitempty"`
	StripeInvoiceID *string    `json:"stripe_invoice_id,omitempty"`
	PeriodStart     time.Time  `json:"period_start"`
	PeriodEnd       time.Time  `json:"period_end"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type SubscriptionRepository struct {
	pool *db.Pool
}

func NewSubscriptionRepository(pool *db.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) Create(ctx context.Context, s *Subscription) error {
	query := `INSERT INTO subscriptions (user_id, plan_id, status, current_period_start, current_period_end, cancelled_at, trial_end, stripe_subscription_id, metadata)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	          RETURNING id, created_at, updated_at`
	metadataJSON, _ := json.Marshal(s.Metadata)
	return r.pool.QueryRow(ctx, query,
		s.UserID, s.PlanID, s.Status, s.CurrentPeriodStart, s.CurrentPeriodEnd,
		s.CancelledAt, s.TrialEnd, s.StripeSubscriptionID, metadataJSON,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*Subscription, error) {
	query := `SELECT id, user_id, plan_id, status, current_period_start, current_period_end,
	                 cancelled_at, trial_end, stripe_subscription_id, metadata, created_at, updated_at
	          FROM subscriptions WHERE id = $1 AND deleted_at IS NULL`
	s := &Subscription{}
	var cancelledAt, trialEnd *time.Time
	var stripeSubID *string
	var metadataJSON []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.PlanID, &s.Status,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd,
		&cancelledAt, &trialEnd, &stripeSubID,
		&metadataJSON, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	s.CancelledAt = cancelledAt
	s.TrialEnd = trialEnd
	s.StripeSubscriptionID = stripeSubID
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &s.Metadata)
	}
	return s, nil
}

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Subscription, error) {
	query := `SELECT id, user_id, plan_id, status, current_period_start, current_period_end,
	                 cancelled_at, trial_end, stripe_subscription_id, metadata, created_at, updated_at
	          FROM subscriptions WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*Subscription
	for rows.Next() {
		s := &Subscription{}
		var cancelledAt, trialEnd *time.Time
		var stripeSubID *string
		var metadataJSON []byte
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.PlanID, &s.Status,
			&s.CurrentPeriodStart, &s.CurrentPeriodEnd,
			&cancelledAt, &trialEnd, &stripeSubID,
			&metadataJSON, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		s.CancelledAt = cancelledAt
		s.TrialEnd = trialEnd
		s.StripeSubscriptionID = stripeSubID
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &s.Metadata)
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, s *Subscription) error {
	query := `UPDATE subscriptions SET plan_id = $1, status = $2, current_period_start = $3,
	          current_period_end = $4, cancelled_at = $5, trial_end = $6,
	          stripe_subscription_id = $7, metadata = $8, updated_at = NOW()
	          WHERE id = $9 AND deleted_at IS NULL`
	metadataJSON, _ := json.Marshal(s.Metadata)
	_, err := r.pool.Exec(ctx, query,
		s.PlanID, s.Status, s.CurrentPeriodStart, s.CurrentPeriodEnd,
		s.CancelledAt, s.TrialEnd, s.StripeSubscriptionID, metadataJSON, s.ID,
	)
	return err
}

func (r *SubscriptionRepository) List(ctx context.Context) ([]*Subscription, error) {
	query := `SELECT id, user_id, plan_id, status, current_period_start, current_period_end,
	                 cancelled_at, trial_end, stripe_subscription_id, metadata, created_at, updated_at
	          FROM subscriptions WHERE deleted_at IS NULL ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*Subscription
	for rows.Next() {
		s := &Subscription{}
		var cancelledAt, trialEnd *time.Time
		var stripeSubID *string
		var metadataJSON []byte
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.PlanID, &s.Status,
			&s.CurrentPeriodStart, &s.CurrentPeriodEnd,
			&cancelledAt, &trialEnd, &stripeSubID,
			&metadataJSON, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		s.CancelledAt = cancelledAt
		s.TrialEnd = trialEnd
		s.StripeSubscriptionID = stripeSubID
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &s.Metadata)
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *SubscriptionRepository) GetActiveSubscription(ctx context.Context, userID uuid.UUID) (*Subscription, error) {
	query := `SELECT id, user_id, plan_id, status, current_period_start, current_period_end,
	                 cancelled_at, trial_end, stripe_subscription_id, metadata, created_at, updated_at
	          FROM subscriptions WHERE user_id = $1 AND status = 'active' AND deleted_at IS NULL
	          ORDER BY current_period_end DESC LIMIT 1`
	s := &Subscription{}
	var cancelledAt, trialEnd *time.Time
	var stripeSubID *string
	var metadataJSON []byte
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.PlanID, &s.Status,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd,
		&cancelledAt, &trialEnd, &stripeSubID,
		&metadataJSON, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	s.CancelledAt = cancelledAt
	s.TrialEnd = trialEnd
	s.StripeSubscriptionID = stripeSubID
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &s.Metadata)
	}
	return s, nil
}

func (r *SubscriptionRepository) ListPlans(ctx context.Context) ([]*SubscriptionPlan, error) {
	query := `SELECT id, name, description, amount, currency, interval, features, is_active, created_at
	          FROM subscription_plans WHERE is_active = TRUE ORDER BY amount`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*SubscriptionPlan
	for rows.Next() {
		p := &SubscriptionPlan{}
		var featuresJSON []byte
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Amount, &p.Currency,
			&p.Interval, &featuresJSON, &p.IsActive, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		if featuresJSON != nil {
			json.Unmarshal(featuresJSON, &p.Features)
		}
		plans = append(plans, p)
	}
	return plans, nil
}

func (r *SubscriptionRepository) GetPlanByID(ctx context.Context, id uuid.UUID) (*SubscriptionPlan, error) {
	query := `SELECT id, name, description, amount, currency, interval, features, is_active, created_at
	          FROM subscription_plans WHERE id = $1`
	p := &SubscriptionPlan{}
	var featuresJSON []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Amount, &p.Currency,
		&p.Interval, &featuresJSON, &p.IsActive, &p.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if featuresJSON != nil {
		json.Unmarshal(featuresJSON, &p.Features)
	}
	return p, nil
}

func (r *SubscriptionRepository) CreateBillingRecord(ctx context.Context, b *BillingRecord) error {
	query := `INSERT INTO billing_history (user_id, subscription_id, amount, currency, status, payment_method, stripe_invoice_id, period_start, period_end, paid_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	          RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query,
		b.UserID, b.SubscriptionID, b.Amount, b.Currency, b.Status,
		b.PaymentMethod, b.StripeInvoiceID, b.PeriodStart, b.PeriodEnd, b.PaidAt,
	).Scan(&b.ID, &b.CreatedAt)
}

func (r *SubscriptionRepository) ListBillingByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*BillingRecord, error) {
	query := `SELECT id, user_id, subscription_id, amount, currency, status, payment_method,
	                 stripe_invoice_id, period_start, period_end, paid_at, created_at
	          FROM billing_history WHERE user_id = $1 AND deleted_at IS NULL
	          ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*BillingRecord
	for rows.Next() {
		b := &BillingRecord{}
		var subID *uuid.UUID
		var invoiceID *string
		var paidAt *time.Time
		if err := rows.Scan(
			&b.ID, &b.UserID, &subID, &b.Amount, &b.Currency, &b.Status,
			&b.PaymentMethod, &invoiceID, &b.PeriodStart, &b.PeriodEnd,
			&paidAt, &b.CreatedAt,
		); err != nil {
			return nil, err
		}
		b.SubscriptionID = subID
		b.StripeInvoiceID = invoiceID
		b.PaidAt = paidAt
		records = append(records, b)
	}
	return records, nil
}

func (r *SubscriptionRepository) ListDueForRenewal(ctx context.Context, now time.Time) ([]*Subscription, error) {
	query := `SELECT id, user_id, plan_id, status, current_period_start, current_period_end,
	                 cancelled_at, trial_end, stripe_subscription_id, metadata, created_at, updated_at
	          FROM subscriptions WHERE status = 'active' AND current_period_end <= $1 AND deleted_at IS NULL`
	rows, err := r.pool.Query(ctx, query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*Subscription
	for rows.Next() {
		s := &Subscription{}
		var cancelledAt, trialEnd *time.Time
		var stripeSubID *string
		var metadataJSON []byte
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.PlanID, &s.Status,
			&s.CurrentPeriodStart, &s.CurrentPeriodEnd,
			&cancelledAt, &trialEnd, &stripeSubID,
			&metadataJSON, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		s.CancelledAt = cancelledAt
		s.TrialEnd = trialEnd
		s.StripeSubscriptionID = stripeSubID
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &s.Metadata)
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *SubscriptionRepository) UpdateBillingStatus(ctx context.Context, id uuid.UUID, status string, paidAt *time.Time) error {
	query := `UPDATE billing_history SET status = $1, paid_at = COALESCE($2, paid_at) WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, paidAt, id)
	return err
}
