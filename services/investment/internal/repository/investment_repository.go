// Copyright (c) 2026 RoundPenny. All rights reserved.

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/db"
)

type Portfolio struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Strategy  string    `json:"strategy"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

type Investment struct {
	ID          uuid.UUID `json:"id"`
	PortfolioID uuid.UUID `json:"portfolio_id"`
	UserID      uuid.UUID `json:"user_id"`
	Amount      float64   `json:"amount"`
	Source      string    `json:"source"`
	Status      string    `json:"status"`
	ExternalRef string    `json:"external_ref,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type InvestmentRepository struct {
	pool *db.Pool
}

func NewInvestmentRepository(pool *db.Pool) *InvestmentRepository {
	return &InvestmentRepository{pool: pool}
}

func (r *InvestmentRepository) GetOrCreatePortfolio(ctx context.Context, userID uuid.UUID, strategy string) (*Portfolio, error) {
	var p Portfolio
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, name, strategy, balance, created_at
		FROM portfolios WHERE user_id = $1`, userID,
	).Scan(&p.ID, &p.UserID, &p.Name, &p.Strategy, &p.Balance, &p.CreatedAt)

	if err == nil {
		return &p, nil
	}

	err = r.pool.QueryRow(ctx, `
		INSERT INTO portfolios (user_id, name, strategy)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, name, strategy, balance, created_at`,
		userID, "Round-Up Portfolio", strategy,
	).Scan(&p.ID, &p.UserID, &p.Name, &p.Strategy, &p.Balance, &p.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *InvestmentRepository) CreateInvestment(ctx context.Context, inv *Investment) error {
	query := `
		INSERT INTO investments (portfolio_id, user_id, amount, source, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query,
		inv.PortfolioID, inv.UserID, inv.Amount, inv.Source, inv.Status,
	).Scan(&inv.ID, &inv.CreatedAt)
}

func (r *InvestmentRepository) AddToPortfolioBalance(ctx context.Context, portfolioID uuid.UUID, amount float64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE portfolios SET balance = balance + $1 WHERE id = $2`,
		amount, portfolioID,
	)
	return err
}
