// Copyright (c) 2026 RoundPenny. All rights reserved.

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/roundup-platform/pkg/db"
)

type RoundUpRecord struct {
	ID              uuid.UUID       `json:"id"`
	TransactionID   uuid.UUID       `json:"transaction_id"`
	UserID          uuid.UUID       `json:"user_id"`
	OriginalAmount  float64         `json:"original_amount"`
	RoundedAmount   float64         `json:"rounded_amount"`
	RoundUpAmount   float64         `json:"round_up_amount"`
	Currency        string          `json:"currency"`
	Status          string          `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
}

type UserPreference struct {
	RoundToNearest   float64 `json:"round_to_nearest"`
	MaxDailyRoundup  float64 `json:"max_daily_roundup"`
	Multiplier       int     `json:"multiplier"`
	AutoInvest       bool    `json:"auto_invest"`
}

type RoundUpRepository struct {
	pool *db.Pool
}

func NewRoundUpRepository(pool *db.Pool) *RoundUpRepository {
	return &RoundUpRepository{pool: pool}
}

func (r *RoundUpRepository) GetUserPreferences(ctx context.Context, userID uuid.UUID) (*UserPreference, error) {
	query := `
		SELECT round_to_nearest, max_daily_roundup, multiplier, auto_invest
		FROM user_preferences WHERE user_id = $1`
	pref := &UserPreference{}
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&pref.RoundToNearest, &pref.MaxDailyRoundup, &pref.Multiplier, &pref.AutoInvest,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &UserPreference{
				RoundToNearest:  1.00,
				MaxDailyRoundup: 5.00,
				Multiplier:      1,
				AutoInvest:      true,
			}, nil
		}
		return nil, err
	}
	return pref, nil
}

func (r *RoundUpRepository) GetDailyRoundUpTotal(ctx context.Context, userID uuid.UUID) (float64, error) {
	query := `
		SELECT COALESCE(SUM(round_up_amount), 0)
		FROM transaction_roundups
		WHERE user_id = $1 AND created_at >= CURRENT_DATE AND status != 'failed'`
	var total float64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&total)
	return total, err
}

func (r *RoundUpRepository) CreateRoundUp(ctx context.Context, record *RoundUpRecord) error {
	query := `
		INSERT INTO transaction_roundups
			(transaction_id, user_id, original_amount, rounded_amount, round_up_amount, currency, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query,
		record.TransactionID, record.UserID, record.OriginalAmount,
		record.RoundedAmount, record.RoundUpAmount, record.Currency, record.Status,
	).Scan(&record.ID, &record.CreatedAt)
}

func (r *RoundUpRepository) UpdateRoundUpStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE transaction_roundups SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	return err
}
