package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/db"
)

type JournalEntry struct {
	ID            uuid.UUID `json:"id"`
	TransactionID *uuid.UUID `json:"transaction_id,omitempty"`
	RoundUpID     *uuid.UUID `json:"roundup_id,omitempty"`
	Description   string    `json:"description"`
	EntryDate     time.Time `json:"entry_date"`
	CreatedAt     time.Time `json:"created_at"`
}

type JournalLine struct {
	ID             uuid.UUID `json:"id"`
	JournalEntryID uuid.UUID `json:"journal_entry_id"`
	AccountCode    string    `json:"account_code"`
	DebitAmount    float64   `json:"debit_amount"`
	CreditAmount   float64   `json:"credit_amount"`
	Currency       string    `json:"currency"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
}

type LedgerRepository struct {
	pool *db.Pool
}

func NewLedgerRepository(pool *db.Pool) *LedgerRepository {
	return &LedgerRepository{pool: pool}
}

func (r *LedgerRepository) CreateDoubleEntry(ctx context.Context, description string, txID, roundUpID *uuid.UUID, lines []JournalLine) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var entryID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO journal_entries (transaction_id, roundup_id, description, entry_date)
		VALUES ($1, $2, $3, CURRENT_DATE)
		RETURNING id`,
		txID, roundUpID, description,
	).Scan(&entryID)
	if err != nil {
		return err
	}

	for _, line := range lines {
		_, err = tx.Exec(ctx, `
			INSERT INTO journal_lines (journal_entry_id, account_id, debit_amount, credit_amount, currency, user_id)
			VALUES ($1, (SELECT id FROM ledger_accounts WHERE code = $2), $3, $4, $5, $6)`,
			entryID, line.AccountCode, line.DebitAmount, line.CreditAmount, line.Currency, line.UserID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *LedgerRepository) GetEntry(ctx context.Context, id uuid.UUID) (*JournalEntry, error) {
	return nil, nil
}

func (r *LedgerRepository) ListByAccount(ctx context.Context, accountCode string, limit, offset int) ([]JournalLine, error) {
	return nil, nil
}

func (r *LedgerRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]JournalLine, error) {
	return nil, nil
}
