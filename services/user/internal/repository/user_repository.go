// Copyright (c) 2026 RoundPenny. All rights reserved.

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/db"
)

type UserProfile struct {
	UserID    uuid.UUID `json:"user_id"`
	FullName  string    `json:"full_name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserPreferences struct {
	UserID           uuid.UUID `json:"user_id"`
	RoundToNearest   float64   `json:"round_to_nearest"`
	MaxDailyRoundUp  float64   `json:"max_daily_roundup"`
	Multiplier       int       `json:"multiplier"`
	AutoInvest       bool      `json:"auto_invest"`
	InvestmentStrategy string  `json:"investment_strategy"`
	Language         string    `json:"language"`
	Timezone         string    `json:"timezone"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UserRepository struct {
	pool *db.Pool
}

func NewUserRepository(pool *db.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	query := `
		SELECT u.full_name, u.phone, u.created_at, u.updated_at
		FROM users u WHERE u.id = $1 AND u.deleted_at IS NULL`
	p := &UserProfile{UserID: userID}
	err := r.pool.QueryRow(ctx, query, userID).Scan(&p.FullName, &p.Phone, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, fullName string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET full_name = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`,
		fullName, userID)
	return err
}

func (r *UserRepository) GetPreferences(ctx context.Context, userID uuid.UUID) (*UserPreferences, error) {
	query := `
		SELECT user_id, round_to_nearest, max_daily_roundup, multiplier, auto_invest,
		       investment_strategy, language, timezone, created_at, updated_at
		FROM user_preferences WHERE user_id = $1`
	p := &UserPreferences{}
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&p.UserID, &p.RoundToNearest, &p.MaxDailyRoundUp, &p.Multiplier, &p.AutoInvest,
		&p.InvestmentStrategy, &p.Language, &p.Timezone, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *UserRepository) UpsertPreferences(ctx context.Context, prefs *UserPreferences) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_preferences (user_id, round_to_nearest, max_daily_roundup, multiplier, auto_invest, investment_strategy, language, timezone)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) DO UPDATE SET
			round_to_nearest = EXCLUDED.round_to_nearest,
			max_daily_roundup = EXCLUDED.max_daily_roundup,
			multiplier = EXCLUDED.multiplier,
			auto_invest = EXCLUDED.auto_invest,
			investment_strategy = EXCLUDED.investment_strategy,
			language = EXCLUDED.language,
			timezone = EXCLUDED.timezone,
			updated_at = NOW()`,
		prefs.UserID, prefs.RoundToNearest, prefs.MaxDailyRoundUp, prefs.Multiplier,
		prefs.AutoInvest, prefs.InvestmentStrategy, prefs.Language, prefs.Timezone)
	return err
}
