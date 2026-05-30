// Copyright (c) 2026 RoundPenny. All rights reserved.

package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/roundup-platform/pkg/db"
)

type User struct {
	ID                  uuid.UUID  `json:"id"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"`
	FullName            string     `json:"full_name"`
	Phone               *string    `json:"phone,omitempty"`
	EmailVerified       bool       `json:"email_verified"`
	KYCStatus           string     `json:"kyc_status"`
	MFAEnabled          bool       `json:"mfa_enabled"`
	MFASecret           *string    `json:"-"`
	MFABackupCodes      []string   `json:"-"`
	FailedLoginAttempts int        `json:"-"`
	LockedUntil         *time.Time `json:"-"`
	Role                string     `json:"role"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type OAuthAccount struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Provider       string
	ProviderUserID string
	AccessToken    string
	RefreshToken   *string
	ExpiresAt      *time.Time
	CreatedAt      time.Time
}

type EmailVerification struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Email      string
	Token      string
	ExpiresAt  time.Time
	VerifiedAt *time.Time
	CreatedAt  time.Time
}

type KYCSubmission struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	FullName        string
	DocumentType    string
	DocumentNumber  string
	Status          string
	RejectionReason *string
	SubmittedAt     time.Time
	ReviewedAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type AuthRepository struct {
	pool *db.Pool
}

func NewAuthRepository(pool *db.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

func (r *AuthRepository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (email, password_hash, full_name, phone)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email_verified, kyc_status, role, created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		user.Email, user.PasswordHash, user.FullName, user.Phone,
	).Scan(&user.ID, &user.EmailVerified, &user.KYCStatus, &user.Role, &user.CreatedAt, &user.UpdatedAt)
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, full_name, phone, email_verified,
		       kyc_status, mfa_enabled, mfa_secret, mfa_backup_codes,
		       failed_login_attempts, locked_until, role, created_at, updated_at
		FROM users WHERE email = $1 AND deleted_at IS NULL`
	user := &User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Phone,
		&user.EmailVerified, &user.KYCStatus, &user.MFAEnabled, &user.MFASecret,
		&user.MFABackupCodes, &user.FailedLoginAttempts, &user.LockedUntil,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("query user: %w", err)
	}
	return user, nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, email, password_hash, full_name, phone, email_verified,
		       kyc_status, mfa_enabled, mfa_secret, mfa_backup_codes,
		       failed_login_attempts, locked_until, role, created_at, updated_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`
	user := &User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Phone,
		&user.EmailVerified, &user.KYCStatus, &user.MFAEnabled, &user.MFASecret,
		&user.MFABackupCodes, &user.FailedLoginAttempts, &user.LockedUntil,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("query user: %w", err)
	}
	return user, nil
}

func (r *AuthRepository) SaveRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, query, userID, tokenHash, expiresAt)
	return err
}

func (r *AuthRepository) GetRefreshToken(ctx context.Context, tokenHash string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM refresh_tokens WHERE token_hash = $1 AND revoked = FALSE AND expires_at > NOW())`
	var exists bool
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(&exists)
	return exists, err
}

func (r *AuthRepository) SaveMFASecret(ctx context.Context, userID uuid.UUID, secret string) error {
	query := `UPDATE users SET mfa_secret = $1, mfa_enabled = FALSE WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, secret, userID)
	return err
}

func (r *AuthRepository) EnableMFA(ctx context.Context, userID uuid.UUID, backupCodes []string) error {
	query := `UPDATE users SET mfa_enabled = TRUE, mfa_backup_codes = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, backupCodes, userID)
	return err
}

func (r *AuthRepository) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET mfa_enabled = FALSE, mfa_secret = NULL, mfa_backup_codes = '{}' WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *AuthRepository) IncrementFailedLogins(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET failed_login_attempts = failed_login_attempts + 1 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *AuthRepository) ResetFailedLogins(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET failed_login_attempts = 0, locked_until = NULL WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *AuthRepository) LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error {
	query := `UPDATE users SET locked_until = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, until, userID)
	return err
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE token_hash = $1`
	_, err := r.pool.Exec(ctx, query, tokenHash)
	return err
}

func (r *AuthRepository) RevokeUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1 AND revoked = FALSE`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *AuthRepository) UpsertOAuthAccount(ctx context.Context, userID uuid.UUID, provider, providerUserID, accessToken, refreshToken string, expiresAt *time.Time) error {
	query := `
		INSERT INTO oauth_accounts (user_id, provider, provider_user_id, access_token, refresh_token, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (provider, provider_user_id) DO UPDATE
		SET access_token = $4, refresh_token = $5, expires_at = $6`
	_, err := r.pool.Exec(ctx, query, userID, provider, providerUserID, accessToken, refreshToken, expiresAt)
	return err
}

func (r *AuthRepository) CreateEmailVerification(ctx context.Context, userID uuid.UUID, email, token string, expiresAt time.Time) error {
	query := `INSERT INTO email_verifications (user_id, email, token, expires_at) VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, query, userID, email, token, expiresAt)
	return err
}

func (r *AuthRepository) GetEmailVerificationByToken(ctx context.Context, token string) (*EmailVerification, error) {
	query := `SELECT id, user_id, email, token, expires_at, verified_at, created_at FROM email_verifications WHERE token = $1`
	v := &EmailVerification{}
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&v.ID, &v.UserID, &v.Email, &v.Token, &v.ExpiresAt, &v.VerifiedAt, &v.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *AuthRepository) VerifyEmail(ctx context.Context, token string) error {
	query := `UPDATE email_verifications SET verified_at = NOW() WHERE token = $1 AND verified_at IS NULL AND expires_at > NOW()`
	_, err := r.pool.Exec(ctx, query, token)
	return err
}

func (r *AuthRepository) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET email_verified = TRUE WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *AuthRepository) CreateKYCSubmission(ctx context.Context, sub *KYCSubmission) error {
	query := `
		INSERT INTO kyc_submissions (user_id, full_name, document_type, document_number)
		VALUES ($1, $2, $3, $4)
		RETURNING id, status, submitted_at, created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		sub.UserID, sub.FullName, sub.DocumentType, sub.DocumentNumber,
	).Scan(&sub.ID, &sub.Status, &sub.SubmittedAt, &sub.CreatedAt, &sub.UpdatedAt)
}

func (r *AuthRepository) GetKYCSubmissionByUserID(ctx context.Context, userID uuid.UUID) (*KYCSubmission, error) {
	query := `SELECT id, user_id, full_name, document_type, document_number, status, rejection_reason, submitted_at, reviewed_at, created_at, updated_at FROM kyc_submissions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
	sub := &KYCSubmission{}
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&sub.ID, &sub.UserID, &sub.FullName, &sub.DocumentType, &sub.DocumentNumber,
		&sub.Status, &sub.RejectionReason, &sub.SubmittedAt, &sub.ReviewedAt,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func (r *AuthRepository) UpdateKYCStatus(ctx context.Context, userID uuid.UUID, status string) error {
	query := `UPDATE users SET kyc_status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, userID)
	return err
}

func GenerateBackupCodes() ([]string, error) {
	codes := make([]string, 8)
	for i := range codes {
		b := make([]byte, 4)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		code := hex.EncodeToString(b)
		codes[i] = code[:4] + "-" + code[4:]
	}
	return codes, nil
}

func (r *AuthRepository) GetOAuthAccount(ctx context.Context, provider, providerUserID string) (*OAuthAccount, error) {
	query := `SELECT id, user_id, provider, provider_user_id, access_token, refresh_token, expires_at, created_at FROM oauth_accounts WHERE provider = $1 AND provider_user_id = $2`
	acc := &OAuthAccount{}
	err := r.pool.QueryRow(ctx, query, provider, providerUserID).Scan(
		&acc.ID, &acc.UserID, &acc.Provider, &acc.ProviderUserID, &acc.AccessToken, &acc.RefreshToken, &acc.ExpiresAt, &acc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return acc, nil
}
