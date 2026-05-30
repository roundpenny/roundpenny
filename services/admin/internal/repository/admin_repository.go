// Copyright (c) 2026 RoundPenny. All rights reserved.

package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Pool defines the database methods used by AdminRepository.
type Pool interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type AdminStats struct {
	TotalUsers        int `json:"total_users"`
	PendingKYC        int `json:"pending_kyc"`
	ActiveMerchants   int `json:"active_merchants"`
	TotalTransactions int `json:"total_transactions"`
	OpenFraudAlerts   int `json:"open_fraud_alerts"`
	PendingPayments   int `json:"pending_payments"`
}

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	FullName      string    `json:"full_name"`
	Phone         string    `json:"phone"`
	EmailVerified bool      `json:"email_verified"`
	KYCStatus     string    `json:"kyc_status"`
	MFaEnabled    bool      `json:"mfa_enabled"`
	Role          string    `json:"role"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Merchant struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	Status        string    `json:"status"`
	FeePercentage float64   `json:"fee_percentage"`
	Country       string    `json:"country"`
	CreatedAt     time.Time `json:"created_at"`
}

type Transaction struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type FraudAlert struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type KYCSubmission struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	FullName        string     `json:"full_name"`
	DocumentType    string     `json:"document_type"`
	DocumentNumber  string     `json:"document_number"`
	Status          string     `json:"status"`
	RejectionReason string     `json:"rejection_reason,omitempty"`
	SubmittedAt     time.Time  `json:"submitted_at"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
}

type AdminRepository struct {
	pool Pool
}

func NewAdminRepository(pool Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

func (r *AdminRepository) GetStats(ctx context.Context) (*AdminStats, error) {
	stats := &AdminStats{}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&stats.TotalUsers); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM kyc_submissions WHERE status = 'pending'`).Scan(&stats.PendingKYC); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM merchants WHERE status = 'active' AND deleted_at IS NULL`).Scan(&stats.ActiveMerchants); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM transactions`).Scan(&stats.TotalTransactions); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM fraud_alerts WHERE status = 'open'`).Scan(&stats.OpenFraudAlerts); err != nil {
		return nil, err
	}
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM payments WHERE status = 'pending'`).Scan(&stats.PendingPayments); err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *AdminRepository) ListUsers(ctx context.Context, offset, limit int) ([]User, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, email, full_name, COALESCE(phone,''), email_verified, 
		       COALESCE(kyc_status,'pending'), mfa_enabled, role, created_at, updated_at
		FROM users WHERE deleted_at IS NULL
		ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.Phone, &u.EmailVerified,
			&u.KYCStatus, &u.MFaEnabled, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *AdminRepository) GetUser(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, email, full_name, COALESCE(phone,''), email_verified,
		       COALESCE(kyc_status,'pending'), mfa_enabled, role, created_at, updated_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`, id).Scan(
		&u.ID, &u.Email, &u.FullName, &u.Phone, &u.EmailVerified,
		&u.KYCStatus, &u.MFaEnabled, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *AdminRepository) UpdateUserStatus(ctx context.Context, id, kycStatus string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET kyc_status = $1, updated_at = NOW() WHERE id = $2`, kycStatus, id)
	return err
}

func (r *AdminRepository) ListMerchants(ctx context.Context, offset, limit int) ([]Merchant, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM merchants WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, email, COALESCE(phone,''), status, fee_percentage, COALESCE(country,''), created_at
		FROM merchants WHERE deleted_at IS NULL
		ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var merchants []Merchant
	for rows.Next() {
		var m Merchant
		if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.Status, &m.FeePercentage, &m.Country, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		merchants = append(merchants, m)
	}
	return merchants, total, nil
}

func (r *AdminRepository) GetMerchant(ctx context.Context, id string) (*Merchant, error) {
	m := &Merchant{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, email, COALESCE(phone,''), status, fee_percentage, COALESCE(country,''), created_at
		FROM merchants WHERE id = $1 AND deleted_at IS NULL`, id).Scan(
		&m.ID, &m.Name, &m.Email, &m.Phone, &m.Status, &m.FeePercentage, &m.Country, &m.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return m, nil
}

func (r *AdminRepository) ListTransactions(ctx context.Context, offset, limit int) ([]Transaction, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM transactions`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, amount, currency, status, COALESCE(description,''), created_at
		FROM transactions ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var txs []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Currency, &t.Status, &t.Description, &t.CreatedAt); err != nil {
			return nil, 0, err
		}
		txs = append(txs, t)
	}
	return txs, total, nil
}

func (r *AdminRepository) GetTransaction(ctx context.Context, id string) (*Transaction, error) {
	t := &Transaction{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, amount, currency, status, COALESCE(description,''), created_at
		FROM transactions WHERE id = $1`, id).Scan(
		&t.ID, &t.UserID, &t.Amount, &t.Currency, &t.Status, &t.Description, &t.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}

func (r *AdminRepository) ListFraudAlerts(ctx context.Context, offset, limit int) ([]FraudAlert, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM fraud_alerts`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, severity, status, COALESCE(description,''), created_at
		FROM fraud_alerts ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var alerts []FraudAlert
	for rows.Next() {
		var a FraudAlert
		if err := rows.Scan(&a.ID, &a.UserID, &a.Severity, &a.Status, &a.Description, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		alerts = append(alerts, a)
	}
	return alerts, total, nil
}

func (r *AdminRepository) ReviewFraudAlert(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE fraud_alerts SET status = $1, resolved_at = NOW() WHERE id = $2`, status, id)
	return err
}

func (r *AdminRepository) ListKYCSubmissions(ctx context.Context, offset, limit int) ([]KYCSubmission, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM kyc_submissions`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, full_name, document_type, document_number, status,
		       COALESCE(rejection_reason,''), submitted_at, reviewed_at
		FROM kyc_submissions ORDER BY submitted_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var subs []KYCSubmission
	for rows.Next() {
		var s KYCSubmission
		if err := rows.Scan(&s.ID, &s.UserID, &s.FullName, &s.DocumentType, &s.DocumentNumber,
			&s.Status, &s.RejectionReason, &s.SubmittedAt, &s.ReviewedAt); err != nil {
			return nil, 0, err
		}
		subs = append(subs, s)
	}
	return subs, total, nil
}

func (r *AdminRepository) ReviewKYCSubmission(ctx context.Context, id, status, rejectionReason string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE kyc_submissions SET status = $1, rejection_reason = $2, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $3`, status, rejectionReason, id)
	return err
}
