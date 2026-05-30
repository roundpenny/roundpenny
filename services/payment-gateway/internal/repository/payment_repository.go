// Copyright (c) 2026 RoundPenny. All rights reserved.

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/roundup-platform/pkg/db"
)

type Payment struct {
	ID                    uuid.UUID       `json:"id"`
	UserID                uuid.UUID       `json:"user_id"`
	TransactionID         *uuid.UUID      `json:"transaction_id,omitempty"`
	Amount                float64         `json:"amount"`
	Currency              string          `json:"currency"`
	Status                string          `json:"status"`
	PaymentMethod         string          `json:"payment_method,omitempty"`
	StripePaymentIntentID string          `json:"stripe_payment_intent_id,omitempty"`
	StripePaymentMethodID string          `json:"stripe_payment_method_id,omitempty"`
	Description           string          `json:"description,omitempty"`
	Metadata              map[string]any  `json:"metadata,omitempty"`
	ErrorMessage          string          `json:"error_message,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type PaymentRepository struct {
	pool *db.Pool
}

func NewPaymentRepository(pool *db.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, p *Payment) error {
	query := `INSERT INTO payments (user_id, amount, currency, status, payment_method, stripe_payment_intent_id, stripe_payment_method_id, description, metadata)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	          RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		p.UserID, p.Amount, p.Currency, p.Status, p.PaymentMethod,
		p.StripePaymentIntentID, p.StripePaymentMethodID,
		p.Description, p.Metadata,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*Payment, error) {
	query := `SELECT id, user_id, transaction_id, amount, currency, status, payment_method,
	                 stripe_payment_intent_id, stripe_payment_method_id, description, metadata,
	                 error_message, created_at, updated_at
	          FROM payments WHERE id = $1`
	p := &Payment{}
	var transactionID *uuid.UUID
	var stripeIntentID, stripeMethodID, description, errorMsg *string
	var metadataJSON []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.UserID, &transactionID, &p.Amount, &p.Currency, &p.Status,
		&p.PaymentMethod, &stripeIntentID, &stripeMethodID, &description,
		&metadataJSON, &errorMsg, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if transactionID != nil {
		p.TransactionID = transactionID
	}
	if stripeIntentID != nil {
		p.StripePaymentIntentID = *stripeIntentID
	}
	if stripeMethodID != nil {
		p.StripePaymentMethodID = *stripeMethodID
	}
	if description != nil {
		p.Description = *description
	}
	if errorMsg != nil {
		p.ErrorMessage = *errorMsg
	}
	return p, nil
}

func (r *PaymentRepository) GetByStripeIntentID(ctx context.Context, stripeIntentID string) (*Payment, error) {
	query := `SELECT id, user_id, transaction_id, amount, currency, status, payment_method,
	                 stripe_payment_intent_id, stripe_payment_method_id, description, metadata,
	                 error_message, created_at, updated_at
	          FROM payments WHERE stripe_payment_intent_id = $1`
	p := &Payment{}
	var transactionID *uuid.UUID
	var stripeMethodID, description, errorMsg *string
	var metadataJSON []byte
	err := r.pool.QueryRow(ctx, query, stripeIntentID).Scan(
		&p.ID, &p.UserID, &transactionID, &p.Amount, &p.Currency, &p.Status,
		&p.PaymentMethod, &p.StripePaymentIntentID, &stripeMethodID, &description,
		&metadataJSON, &errorMsg, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if transactionID != nil {
		p.TransactionID = transactionID
	}
	if stripeMethodID != nil {
		p.StripePaymentMethodID = *stripeMethodID
	}
	if description != nil {
		p.Description = *description
	}
	if errorMsg != nil {
		p.ErrorMessage = *errorMsg
	}
	return p, nil
}

func (r *PaymentRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Payment, error) {
	query := `SELECT id, user_id, transaction_id, amount, currency, status, payment_method,
	                 stripe_payment_intent_id, stripe_payment_method_id, description, metadata,
	                 error_message, created_at, updated_at
	          FROM payments WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*Payment
	for rows.Next() {
		p := &Payment{}
		var transactionID *uuid.UUID
		var stripeIntentID, stripeMethodID, description, errorMsg *string
		var metadataJSON []byte
		if err := rows.Scan(
			&p.ID, &p.UserID, &transactionID, &p.Amount, &p.Currency, &p.Status,
			&p.PaymentMethod, &stripeIntentID, &stripeMethodID, &description,
			&metadataJSON, &errorMsg, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if transactionID != nil {
			p.TransactionID = transactionID
		}
		if stripeIntentID != nil {
			p.StripePaymentIntentID = *stripeIntentID
		}
		if stripeMethodID != nil {
			p.StripePaymentMethodID = *stripeMethodID
		}
		if description != nil {
			p.Description = *description
		}
		if errorMsg != nil {
			p.ErrorMessage = *errorMsg
		}
		payments = append(payments, p)
	}
	return payments, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, transactionID *uuid.UUID, errorMessage string) error {
	query := `UPDATE payments SET status = $1, transaction_id = COALESCE($2, transaction_id),
	          error_message = $3, updated_at = NOW() WHERE id = $4`
	_, err := r.pool.Exec(ctx, query, status, transactionID, errorMessage, id)
	return err
}

func (r *PaymentRepository) UpdateStripePaymentMethod(ctx context.Context, id uuid.UUID, stripePaymentMethodID string) error {
	query := `UPDATE payments SET stripe_payment_method_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, stripePaymentMethodID, id)
	return err
}

func (r *PaymentRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*Payment, error) {
	query := `SELECT id, user_id, transaction_id, amount, currency, status, payment_method,
	                 stripe_payment_intent_id, stripe_payment_method_id, description, metadata,
	                 error_message, created_at, updated_at
	          FROM payments WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*Payment
	for rows.Next() {
		p := &Payment{}
		var transactionID *uuid.UUID
		var stripeIntentID, stripeMethodID, description, errorMsg *string
		var metadataJSON []byte
		if err := rows.Scan(
			&p.ID, &p.UserID, &transactionID, &p.Amount, &p.Currency, &p.Status,
			&p.PaymentMethod, &stripeIntentID, &stripeMethodID, &description,
			&metadataJSON, &errorMsg, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if transactionID != nil {
			p.TransactionID = transactionID
		}
		if stripeIntentID != nil {
			p.StripePaymentIntentID = *stripeIntentID
		}
		if stripeMethodID != nil {
			p.StripePaymentMethodID = *stripeMethodID
		}
		if description != nil {
			p.Description = *description
		}
		if errorMsg != nil {
			p.ErrorMessage = *errorMsg
		}
		payments = append(payments, p)
	}
	return payments, nil
}
