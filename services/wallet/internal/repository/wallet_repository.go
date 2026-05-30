// Copyright (c) 2026 RoundPenny. All rights reserved.

package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/roundup-platform/pkg/db"
)

type Wallet struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	Version   int       `json:"version"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type WalletEntry struct {
	ID            uuid.UUID `json:"id"`
	WalletID      uuid.UUID `json:"wallet_id"`
	UserID        uuid.UUID `json:"user_id"`
	Amount        float64   `json:"amount"`
	BalanceBefore float64   `json:"balance_before"`
	BalanceAfter  float64   `json:"balance_after"`
	EntryType     string    `json:"entry_type"`
	ReferenceType string    `json:"reference_type,omitempty"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

type WalletRepository struct {
	pool *db.Pool
}

func NewWalletRepository(pool *db.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

func (r *WalletRepository) GetOrCreate(ctx context.Context, userID uuid.UUID) (*Wallet, error) {
	var w Wallet
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, balance, currency, version, status, created_at
		FROM wallets WHERE user_id = $1`, userID,
	).Scan(&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.Version, &w.Status, &w.CreatedAt)

	if err == nil {
		return &w, nil
	}

	if err != pgx.ErrNoRows {
		return nil, err
	}

	err = r.pool.QueryRow(ctx, `
		INSERT INTO wallets (user_id, balance, currency)
		VALUES ($1, 0, 'USD')
		RETURNING id, user_id, balance, currency, version, status, created_at`,
		userID,
	).Scan(&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.Version, &w.Status, &w.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &w, nil
}

func (r *WalletRepository) Credit(ctx context.Context, walletID uuid.UUID, amount float64, entryType, refType, description string, refID *uuid.UUID) (*WalletEntry, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var w Wallet
	err = tx.QueryRow(ctx, `
		SELECT id, user_id, balance, currency, version, status
		FROM wallets WHERE id = $1 FOR UPDATE`, walletID,
	).Scan(&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.Version, &w.Status)

	if err != nil {
		return nil, fmt.Errorf("lock wallet: %w", err)
	}

	newBalance := w.Balance + amount

	result, err := tx.Exec(ctx, `
		UPDATE wallets SET balance = $1, version = version + 1, updated_at = NOW()
		WHERE id = $2 AND version = $3`,
		newBalance, walletID, w.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("update wallet: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("concurrent modification detected")
	}

	entry := &WalletEntry{
		WalletID:      walletID,
		UserID:        w.UserID,
		Amount:        amount,
		BalanceBefore: w.Balance,
		BalanceAfter:  newBalance,
		EntryType:     entryType,
		ReferenceType: refType,
		ReferenceID:   refID,
		Description:   description,
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO wallet_entries (wallet_id, user_id, amount, balance_before, balance_after, entry_type, reference_type, reference_id, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`,
		entry.WalletID, entry.UserID, entry.Amount, entry.BalanceBefore,
		entry.BalanceAfter, entry.EntryType, entry.ReferenceType,
		entry.ReferenceID, entry.Description,
	).Scan(&entry.ID, &entry.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("create entry: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return entry, nil
}

func (r *WalletRepository) GetEntries(ctx context.Context, userID uuid.UUID, limit, offset int) ([]WalletEntry, error) {
	query := `
		SELECT we.id, we.wallet_id, we.user_id, we.amount, we.balance_before, we.balance_after,
		       we.entry_type, we.reference_type, we.reference_id, we.description, we.created_at
		FROM wallet_entries we
		WHERE we.user_id = $1
		ORDER BY we.created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []WalletEntry
	for rows.Next() {
		var e WalletEntry
		if err := rows.Scan(
			&e.ID, &e.WalletID, &e.UserID, &e.Amount, &e.BalanceBefore,
			&e.BalanceAfter, &e.EntryType, &e.ReferenceType, &e.ReferenceID,
			&e.Description, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
