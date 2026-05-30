package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/roundup-platform/services/admin/internal/repository"
)

// AdminRepository defines the data access methods needed by AdminService.
type AdminRepository interface {
	GetStats(ctx context.Context) (*repository.AdminStats, error)
	ListUsers(ctx context.Context, offset, limit int) ([]repository.User, int, error)
	GetUser(ctx context.Context, id string) (*repository.User, error)
	UpdateUserStatus(ctx context.Context, id, kycStatus string) error
	ListMerchants(ctx context.Context, offset, limit int) ([]repository.Merchant, int, error)
	GetMerchant(ctx context.Context, id string) (*repository.Merchant, error)
	ListTransactions(ctx context.Context, offset, limit int) ([]repository.Transaction, int, error)
	GetTransaction(ctx context.Context, id string) (*repository.Transaction, error)
	ListFraudAlerts(ctx context.Context, offset, limit int) ([]repository.FraudAlert, int, error)
	ReviewFraudAlert(ctx context.Context, id, status string) error
	ListKYCSubmissions(ctx context.Context, offset, limit int) ([]repository.KYCSubmission, int, error)
	ReviewKYCSubmission(ctx context.Context, id, status, rejectionReason string) error
}

type AdminService struct {
	repo      AdminRepository
	jwtSecret []byte
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotFound           = errors.New("not found")
)

type LoginResponse struct {
	Token     string `json:"access_token"`
	TokenType string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"`
}

func NewAdminService(repo AdminRepository, jwtSecret []byte) *AdminService {
	return &AdminService{repo: repo, jwtSecret: jwtSecret}
}

func (s *AdminService) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	// Simple admin auth: check against env vars
	// In production, use a proper admin users table
	adminEmail := "admin@roundpenny.com"
	adminPass := "admin123"

	if email != adminEmail || password != adminPass {
		return nil, ErrInvalidCredentials
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   "admin",
		"email": email,
		"role":  "admin",
		"iat":   now.Unix(),
		"exp":   now.Add(24 * time.Hour).Unix(),
		"jti":   generateJTI(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:     tokenString,
		TokenType: "Bearer",
		ExpiresIn: 86400,
	}, nil
}

func (s *AdminService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func (s *AdminService) GetStats(ctx context.Context) (*repository.AdminStats, error) {
	return s.repo.GetStats(ctx)
}

func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int) ([]repository.User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListUsers(ctx, offset, pageSize)
}

func (s *AdminService) GetUser(ctx context.Context, id string) (*repository.User, error) {
	return s.repo.GetUser(ctx, id)
}

func (s *AdminService) UpdateUserStatus(ctx context.Context, id, kycStatus string) error {
	return s.repo.UpdateUserStatus(ctx, id, kycStatus)
}

func (s *AdminService) ListMerchants(ctx context.Context, page, pageSize int) ([]repository.Merchant, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListMerchants(ctx, offset, pageSize)
}

func (s *AdminService) GetMerchant(ctx context.Context, id string) (*repository.Merchant, error) {
	return s.repo.GetMerchant(ctx, id)
}

func (s *AdminService) ListTransactions(ctx context.Context, page, pageSize int) ([]repository.Transaction, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListTransactions(ctx, offset, pageSize)
}

func (s *AdminService) GetTransaction(ctx context.Context, id string) (*repository.Transaction, error) {
	return s.repo.GetTransaction(ctx, id)
}

func (s *AdminService) ListFraudAlerts(ctx context.Context, page, pageSize int) ([]repository.FraudAlert, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListFraudAlerts(ctx, offset, pageSize)
}

func (s *AdminService) ReviewFraudAlert(ctx context.Context, id, status string) error {
	valid := map[string]bool{"resolved": true, "false_positive": true, "investigated": true}
	if !valid[status] {
		return errors.New("invalid status: must be resolved, false_positive, or investigated")
	}
	return s.repo.ReviewFraudAlert(ctx, id, status)
}

func (s *AdminService) ListKYCSubmissions(ctx context.Context, page, pageSize int) ([]repository.KYCSubmission, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListKYCSubmissions(ctx, offset, pageSize)
}

func (s *AdminService) ReviewKYCSubmission(ctx context.Context, id, status, rejectionReason string) error {
	valid := map[string]bool{"approved": true, "rejected": true}
	if !valid[status] {
		return errors.New("invalid status: must be approved or rejected")
	}
	if status == "rejected" && rejectionReason == "" {
		return errors.New("rejection reason required")
	}
	reason := ""
	if status == "rejected" {
		reason = rejectionReason
	}
	return s.repo.ReviewKYCSubmission(ctx, id, status, reason)
}

func generateJTI() string {
	h := sha256.Sum256([]byte(time.Now().String()))
	return hex.EncodeToString(h[:16])
}
