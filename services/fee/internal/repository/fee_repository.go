package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/db"
)

var ErrNotFound = errors.New("not found")

type FeeConfig struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	FeeType     string     `json:"fee_type"`
	Percentage  *float64   `json:"percentage,omitempty"`
	FlatAmount  *float64   `json:"flat_amount,omitempty"`
	MinAmount   *float64   `json:"min_amount,omitempty"`
	MaxAmount   *float64   `json:"max_amount,omitempty"`
	IsActive    bool       `json:"is_active"`
}

type FeeTransaction struct {
	ID               uuid.UUID `json:"id"`
	TransactionID    *uuid.UUID `json:"transaction_id,omitempty"`
	RoundUpID        *uuid.UUID `json:"roundup_id,omitempty"`
	UserID           uuid.UUID `json:"user_id"`
	Amount           float64   `json:"amount"`
	FeeType          string    `json:"fee_type"`
	PercentageApplied *float64 `json:"percentage_applied,omitempty"`
	Currency         string    `json:"currency"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

type FeeRepository struct {
	pool *db.Pool
}

func NewFeeRepository(pool *db.Pool) *FeeRepository {
	return &FeeRepository{pool: pool}
}

func (r *FeeRepository) GetActiveConfig(ctx context.Context, feeType string) (*FeeConfig, error) {
	query := `SELECT id, name, fee_type, percentage, flat_amount, min_amount, max_amount, is_active
		FROM fee_configs WHERE fee_type = $1 AND is_active = TRUE LIMIT 1`
	cfg := &FeeConfig{}
	err := r.pool.QueryRow(ctx, query, feeType).Scan(
		&cfg.ID, &cfg.Name, &cfg.FeeType, &cfg.Percentage, &cfg.FlatAmount,
		&cfg.MinAmount, &cfg.MaxAmount, &cfg.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (r *FeeRepository) CreateFeeTransaction(ctx context.Context, ft *FeeTransaction) error {
	query := `
		INSERT INTO fee_transactions (transaction_id, roundup_id, user_id, amount, fee_type, percentage_applied, currency, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query,
		ft.TransactionID, ft.RoundUpID, ft.UserID, ft.Amount, ft.FeeType,
		ft.PercentageApplied, ft.Currency, ft.Status,
	).Scan(&ft.ID, &ft.CreatedAt)
}
