package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/db"
)

type Transaction struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	MerchantID    *uuid.UUID `json:"merchant_id,omitempty"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	Type          string     `json:"type"`
	ExternalTxID  string     `json:"external_tx_id"`
	ExternalProvider string  `json:"external_provider"`
	Description   string     `json:"description"`
	SettledAt     *time.Time `json:"settled_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type TransactionRepository struct {
	pool *db.Pool
}

func NewTransactionRepository(pool *db.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *Transaction) error {
	query := `
		INSERT INTO transactions (user_id, merchant_id, amount, currency, type, external_tx_id, external_provider, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, status, created_at`
	return r.pool.QueryRow(ctx, query,
		tx.UserID, tx.MerchantID, tx.Amount, tx.Currency, tx.Type,
		tx.ExternalTxID, tx.ExternalProvider, tx.Description,
	).Scan(&tx.ID, &tx.Status, &tx.CreatedAt)
}

func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error) {
	query := `
		SELECT id, user_id, merchant_id, amount, currency, status, type,
		       external_tx_id, external_provider, description, settled_at, created_at
		FROM transactions WHERE id = $1`
	tx := &Transaction{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tx.ID, &tx.UserID, &tx.MerchantID, &tx.Amount, &tx.Currency,
		&tx.Status, &tx.Type, &tx.ExternalTxID, &tx.ExternalProvider,
		&tx.Description, &tx.SettledAt, &tx.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (r *TransactionRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Transaction, error) {
	query := `
		SELECT id, user_id, merchant_id, amount, currency, status, type,
		       external_tx_id, external_provider, description, settled_at, created_at
		FROM transactions WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.MerchantID, &tx.Amount, &tx.Currency,
			&tx.Status, &tx.Type, &tx.ExternalTxID, &tx.ExternalProvider,
			&tx.Description, &tx.SettledAt, &tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func (r *TransactionRepository) Settle(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	query := `UPDATE transactions SET status = 'settled', settled_at = $1 WHERE id = $2 AND status = 'pending'`
	_, err := r.pool.Exec(ctx, query, now, id)
	return err
}
