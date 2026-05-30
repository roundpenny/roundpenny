// Copyright (c) 2026 RoundPenny. All rights reserved.

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/db"
)

type Webhook struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	URL         string    `json:"url"`
	Secret      string    `json:"secret,omitempty"`
	Events      []string  `json:"events"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description,omitempty"`
	RetryCount  int       `json:"retry_count"`
	TimeoutMs   int       `json:"timeout_ms"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WebhookDelivery struct {
	ID           uuid.UUID  `json:"id"`
	WebhookID    uuid.UUID  `json:"webhook_id"`
	EventType    string     `json:"event_type"`
	EventID      string     `json:"event_id"`
	Payload      []byte     `json:"payload"`
	Status       string     `json:"status"`
	ResponseCode *int       `json:"response_code,omitempty"`
	ResponseBody string     `json:"response_body,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	AttemptCount int        `json:"attempt_count"`
	MaxAttempts  int        `json:"max_attempts"`
	NextRetryAt  *time.Time `json:"next_retry_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

type WebhookRepository struct {
	pool *db.Pool
}

func NewWebhookRepository(pool *db.Pool) *WebhookRepository {
	return &WebhookRepository{pool: pool}
}

func (r *WebhookRepository) Create(ctx context.Context, w *Webhook) error {
	query := `INSERT INTO webhooks (user_id, url, secret, events, is_active, description, retry_count, timeout_ms)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	          RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		w.UserID, w.URL, w.Secret, w.Events, w.IsActive,
		w.Description, w.RetryCount, w.TimeoutMs,
	).Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)
}

func (r *WebhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*Webhook, error) {
	query := `SELECT id, user_id, url, COALESCE(secret,''), events, is_active,
	                 COALESCE(description,''), retry_count, timeout_ms, created_at, updated_at
	          FROM webhooks WHERE id = $1`
	w := &Webhook{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&w.ID, &w.UserID, &w.URL, &w.Secret, &w.Events, &w.IsActive,
		&w.Description, &w.RetryCount, &w.TimeoutMs, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *WebhookRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Webhook, error) {
	query := `SELECT id, user_id, url, COALESCE(secret,''), events, is_active,
	                 COALESCE(description,''), retry_count, timeout_ms, created_at, updated_at
	          FROM webhooks WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []*Webhook
	for rows.Next() {
		w := &Webhook{}
		if err := rows.Scan(
			&w.ID, &w.UserID, &w.URL, &w.Secret, &w.Events, &w.IsActive,
			&w.Description, &w.RetryCount, &w.TimeoutMs, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, w)
	}
	return webhooks, nil
}

func (r *WebhookRepository) GetActiveByEvent(ctx context.Context, eventType string) ([]*Webhook, error) {
	query := `SELECT id, user_id, url, COALESCE(secret,''), events, is_active,
	                 COALESCE(description,''), retry_count, timeout_ms, created_at, updated_at
	          FROM webhooks WHERE is_active = true AND $1 = ANY(events)`
	rows, err := r.pool.Query(ctx, query, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []*Webhook
	for rows.Next() {
		w := &Webhook{}
		if err := rows.Scan(
			&w.ID, &w.UserID, &w.URL, &w.Secret, &w.Events, &w.IsActive,
			&w.Description, &w.RetryCount, &w.TimeoutMs, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, w)
	}
	return webhooks, nil
}

func (r *WebhookRepository) Update(ctx context.Context, w *Webhook) error {
	query := `UPDATE webhooks SET url=$1, secret=$2, events=$3, is_active=$4,
	          description=$5, retry_count=$6, timeout_ms=$7, updated_at=NOW()
	          WHERE id=$8`
	_, err := r.pool.Exec(ctx, query,
		w.URL, w.Secret, w.Events, w.IsActive,
		w.Description, w.RetryCount, w.TimeoutMs, w.ID,
	)
	return err
}

func (r *WebhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM webhooks WHERE id=$1`, id)
	return err
}

func (r *WebhookRepository) CreateDelivery(ctx context.Context, d *WebhookDelivery) error {
	query := `INSERT INTO webhook_deliveries (webhook_id, event_type, event_id, payload, status, max_attempts, next_retry_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)
	          RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query,
		d.WebhookID, d.EventType, d.EventID, d.Payload, d.Status, d.MaxAttempts, d.NextRetryAt,
	).Scan(&d.ID, &d.CreatedAt)
}

func (r *WebhookRepository) UpdateDelivery(ctx context.Context, id uuid.UUID, status string, responseCode *int, responseBody, errorMessage string, attemptCount int, nextRetryAt *time.Time, completedAt *time.Time) error {
	query := `UPDATE webhook_deliveries SET status=$1, response_code=$2, response_body=$3,
	          error_message=$4, attempt_count=$5, next_retry_at=$6, completed_at=$7
	          WHERE id=$8`
	_, err := r.pool.Exec(ctx, query, status, responseCode, responseBody, errorMessage, attemptCount, nextRetryAt, completedAt, id)
	return err
}
