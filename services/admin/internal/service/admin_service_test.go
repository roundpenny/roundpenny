package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/roundup-platform/services/admin/internal/repository"
)

type mockRepo struct {
	getStatsFn            func(ctx context.Context) (*repository.AdminStats, error)
	listUsersFn           func(ctx context.Context, offset, limit int) ([]repository.User, int, error)
	getUserFn             func(ctx context.Context, id string) (*repository.User, error)
	updateUserStatusFn    func(ctx context.Context, id, kycStatus string) error
	listMerchantsFn       func(ctx context.Context, offset, limit int) ([]repository.Merchant, int, error)
	getMerchantFn         func(ctx context.Context, id string) (*repository.Merchant, error)
	listTransactionsFn    func(ctx context.Context, offset, limit int) ([]repository.Transaction, int, error)
	getTransactionFn      func(ctx context.Context, id string) (*repository.Transaction, error)
	listFraudAlertsFn     func(ctx context.Context, offset, limit int) ([]repository.FraudAlert, int, error)
	reviewFraudAlertFn    func(ctx context.Context, id, status string) error
	listKYCSubmissionsFn  func(ctx context.Context, offset, limit int) ([]repository.KYCSubmission, int, error)
	reviewKYCSubmissionFn func(ctx context.Context, id, status, rejectionReason string) error
}

func (m *mockRepo) GetStats(ctx context.Context) (*repository.AdminStats, error) {
	return m.getStatsFn(ctx)
}
func (m *mockRepo) ListUsers(ctx context.Context, offset, limit int) ([]repository.User, int, error) {
	return m.listUsersFn(ctx, offset, limit)
}
func (m *mockRepo) GetUser(ctx context.Context, id string) (*repository.User, error) {
	return m.getUserFn(ctx, id)
}
func (m *mockRepo) UpdateUserStatus(ctx context.Context, id, kycStatus string) error {
	return m.updateUserStatusFn(ctx, id, kycStatus)
}
func (m *mockRepo) ListMerchants(ctx context.Context, offset, limit int) ([]repository.Merchant, int, error) {
	return m.listMerchantsFn(ctx, offset, limit)
}
func (m *mockRepo) GetMerchant(ctx context.Context, id string) (*repository.Merchant, error) {
	return m.getMerchantFn(ctx, id)
}
func (m *mockRepo) ListTransactions(ctx context.Context, offset, limit int) ([]repository.Transaction, int, error) {
	return m.listTransactionsFn(ctx, offset, limit)
}
func (m *mockRepo) GetTransaction(ctx context.Context, id string) (*repository.Transaction, error) {
	return m.getTransactionFn(ctx, id)
}
func (m *mockRepo) ListFraudAlerts(ctx context.Context, offset, limit int) ([]repository.FraudAlert, int, error) {
	return m.listFraudAlertsFn(ctx, offset, limit)
}
func (m *mockRepo) ReviewFraudAlert(ctx context.Context, id, status string) error {
	return m.reviewFraudAlertFn(ctx, id, status)
}
func (m *mockRepo) ListKYCSubmissions(ctx context.Context, offset, limit int) ([]repository.KYCSubmission, int, error) {
	return m.listKYCSubmissionsFn(ctx, offset, limit)
}
func (m *mockRepo) ReviewKYCSubmission(ctx context.Context, id, status, rejectionReason string) error {
	return m.reviewKYCSubmissionFn(ctx, id, status, rejectionReason)
}

func TestDashboardStats(t *testing.T) {
	svc := NewAdminService(&mockRepo{
		getStatsFn: func(ctx context.Context) (*repository.AdminStats, error) {
			return &repository.AdminStats{
				TotalUsers:        100,
				PendingKYC:        15,
				ActiveMerchants:   42,
				TotalTransactions: 5678,
				OpenFraudAlerts:   3,
				PendingPayments:   8,
			}, nil
		},
	}, []byte("test-secret"))

	stats, err := svc.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalUsers != 100 {
		t.Errorf("TotalUsers = %d, want 100", stats.TotalUsers)
	}
	if stats.PendingKYC != 15 {
		t.Errorf("PendingKYC = %d, want 15", stats.PendingKYC)
	}
	if stats.ActiveMerchants != 42 {
		t.Errorf("ActiveMerchants = %d, want 42", stats.ActiveMerchants)
	}
	if stats.TotalTransactions != 5678 {
		t.Errorf("TotalTransactions = %d, want 5678", stats.TotalTransactions)
	}
	if stats.OpenFraudAlerts != 3 {
		t.Errorf("OpenFraudAlerts = %d, want 3", stats.OpenFraudAlerts)
	}
	if stats.PendingPayments != 8 {
		t.Errorf("PendingPayments = %d, want 8", stats.PendingPayments)
	}
}

func TestDashboardStats_RepoError(t *testing.T) {
	svc := NewAdminService(&mockRepo{
		getStatsFn: func(ctx context.Context) (*repository.AdminStats, error) {
			return nil, errors.New("db error")
		},
	}, []byte("test-secret"))

	_, err := svc.GetStats(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetUserDetail(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	svc := NewAdminService(&mockRepo{
		getUserFn: func(ctx context.Context, id string) (*repository.User, error) {
			return &repository.User{
				ID:            id,
				Email:         "alice@test.com",
				FullName:      "Alice Smith",
				Phone:         "+1234567890",
				EmailVerified: true,
				KYCStatus:     "approved",
				MFaEnabled:    false,
				Role:          "user",
				CreatedAt:     now,
				UpdatedAt:     now.Add(24 * time.Hour),
			}, nil
		},
	}, []byte("test-secret"))

	user, err := svc.GetUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.ID != "user-1" {
		t.Errorf("ID = %q, want user-1", user.ID)
	}
	if user.Email != "alice@test.com" {
		t.Errorf("Email = %q, want alice@test.com", user.Email)
	}
	if user.KYCStatus != "approved" {
		t.Errorf("KYCStatus = %q, want approved", user.KYCStatus)
	}
	if user.Role != "user" {
		t.Errorf("Role = %q, want user", user.Role)
	}
}

func TestGetUserDetail_NotFound(t *testing.T) {
	svc := NewAdminService(&mockRepo{
		getUserFn: func(ctx context.Context, id string) (*repository.User, error) {
			return nil, nil
		},
	}, []byte("test-secret"))

	user, err := svc.GetUser(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != nil {
		t.Fatal("expected nil user, got non-nil")
	}
}

func TestGetUserDetail_RepoError(t *testing.T) {
	svc := NewAdminService(&mockRepo{
		getUserFn: func(ctx context.Context, id string) (*repository.User, error) {
			return nil, errors.New("db error")
		},
	}, []byte("test-secret"))

	_, err := svc.GetUser(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListKycSubmissions(t *testing.T) {
	now := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	svc := NewAdminService(&mockRepo{
		listKYCSubmissionsFn: func(ctx context.Context, offset, limit int) ([]repository.KYCSubmission, int, error) {
			return []repository.KYCSubmission{
				{
					ID:             "kyc-1",
					UserID:         "user-1",
					FullName:       "Bob",
					DocumentType:   "passport",
					DocumentNumber: "AB123456",
					Status:         "pending",
					SubmittedAt:    now,
				},
				{
					ID:             "kyc-2",
					UserID:         "user-2",
					FullName:       "Carol",
					DocumentType:   "license",
					DocumentNumber: "CD789012",
					Status:         "approved",
					SubmittedAt:    now.Add(-24 * time.Hour),
				},
			}, 2, nil
		},
	}, []byte("test-secret"))

	subs, total, err := svc.ListKYCSubmissions(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(subs) != 2 {
		t.Errorf("len(subs) = %d, want 2", len(subs))
	}
	if subs[0].ID != "kyc-1" {
		t.Errorf("subs[0].ID = %q, want kyc-1", subs[0].ID)
	}
	if subs[1].Status != "approved" {
		t.Errorf("subs[1].Status = %q, want approved", subs[1].Status)
	}
}

func TestListKycSubmissions_Pagination(t *testing.T) {
	svc := NewAdminService(&mockRepo{
		listKYCSubmissionsFn: func(ctx context.Context, offset, limit int) ([]repository.KYCSubmission, int, error) {
			if offset != 20 {
				t.Errorf("offset = %d, want 20", offset)
			}
			if limit != 20 {
				t.Errorf("limit = %d, want 20", limit)
			}
			return nil, 0, nil
		},
	}, []byte("test-secret"))

	svc.ListKYCSubmissions(context.Background(), 2, 20)
}

func TestListKycSubmissions_Defaults(t *testing.T) {
	offsetCheck := -1
	limitCheck := -1
	svc := NewAdminService(&mockRepo{
		listKYCSubmissionsFn: func(ctx context.Context, offset, limit int) ([]repository.KYCSubmission, int, error) {
			offsetCheck = offset
			limitCheck = limit
			return nil, 0, nil
		},
	}, []byte("test-secret"))

	svc.ListKYCSubmissions(context.Background(), 0, 0)
	if offsetCheck != 0 {
		t.Errorf("offset = %d, want 0", offsetCheck)
	}
	if limitCheck != 20 {
		t.Errorf("limit = %d, want 20", limitCheck)
	}
}

func TestListKycSubmissions_MaxPageSize(t *testing.T) {
	limitCheck := -1
	svc := NewAdminService(&mockRepo{
		listKYCSubmissionsFn: func(ctx context.Context, offset, limit int) ([]repository.KYCSubmission, int, error) {
			limitCheck = limit
			return nil, 0, nil
		},
	}, []byte("test-secret"))

	svc.ListKYCSubmissions(context.Background(), 1, 200)
	if limitCheck != 20 {
		t.Errorf("limit = %d, want 20 (default when out of range)", limitCheck)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := NewAdminService(&mockRepo{}, []byte("test-secret"))

	resp, err := svc.Login(context.Background(), "admin@roundpenny.com", "admin123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want Bearer", resp.TokenType)
	}
	if resp.ExpiresIn != 86400 {
		t.Errorf("ExpiresIn = %d, want 86400", resp.ExpiresIn)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	svc := NewAdminService(&mockRepo{}, []byte("test-secret"))

	_, err := svc.Login(context.Background(), "wrong@email.com", "password")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := NewAdminService(&mockRepo{}, []byte("test-secret"))

	_, err := svc.Login(context.Background(), "admin@roundpenny.com", "wrong-password")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestValidateToken_Valid(t *testing.T) {
	svc := NewAdminService(&mockRepo{}, []byte("test-secret"))

	resp, err := svc.Login(context.Background(), "admin@roundpenny.com", "admin123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	claims, err := svc.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims["role"] != "admin" {
		t.Errorf("role = %v, want admin", claims["role"])
	}
	if claims["email"] != "admin@roundpenny.com" {
		t.Errorf("email = %v, want admin@roundpenny.com", claims["email"])
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	svc := NewAdminService(&mockRepo{}, []byte("test-secret"))

	_, err := svc.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc := NewAdminService(&mockRepo{}, []byte("test-secret"))

	resp, err := svc.Login(context.Background(), "admin@roundpenny.com", "admin123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	svc2 := NewAdminService(&mockRepo{}, []byte("different-secret"))
	_, err = svc2.ValidateToken(resp.Token)
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	svc := NewAdminService(&mockRepo{}, []byte("test-secret"))

	claims := jwt.MapClaims{
		"sub":   "admin",
		"email": "admin@roundpenny.com",
		"role":  "admin",
		"iat":   time.Now().Add(-2 * time.Hour).Unix(),
		"exp":   time.Now().Add(-1 * time.Hour).Unix(),
		"jti":   generateJTI(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	_, err = svc.ValidateToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestListUsers_Pagination(t *testing.T) {
	capturedOffset := -1
	capturedLimit := -1
	svc := NewAdminService(&mockRepo{
		listUsersFn: func(ctx context.Context, offset, limit int) ([]repository.User, int, error) {
			capturedOffset = offset
			capturedLimit = limit
			return nil, 0, nil
		},
	}, []byte("test-secret"))

	svc.ListUsers(context.Background(), 3, 50)
	if capturedOffset != 100 {
		t.Errorf("offset = %d, want 100 (page 3, pageSize 50)", capturedOffset)
	}
	if capturedLimit != 50 {
		t.Errorf("limit = %d, want 50", capturedLimit)
	}
}

func TestListUsers_DefaultPagination(t *testing.T) {
	capturedOffset := -1
	capturedLimit := -1
	svc := NewAdminService(&mockRepo{
		listUsersFn: func(ctx context.Context, offset, limit int) ([]repository.User, int, error) {
			capturedOffset = offset
			capturedLimit = limit
			return nil, 0, nil
		},
	}, []byte("test-secret"))

	svc.ListUsers(context.Background(), 0, 0)
	if capturedOffset != 0 {
		t.Errorf("offset = %d, want 0", capturedOffset)
	}
	if capturedLimit != 20 {
		t.Errorf("limit = %d, want 20", capturedLimit)
	}
}

func TestReviewFraudAlert_ValidStatus(t *testing.T) {
	capturedStatus := ""
	svc := NewAdminService(&mockRepo{
		reviewFraudAlertFn: func(ctx context.Context, id, status string) error {
			capturedStatus = status
			return nil
		},
	}, []byte("test-secret"))

	err := svc.ReviewFraudAlert(context.Background(), "alert-1", "resolved")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedStatus != "resolved" {
		t.Errorf("status = %q, want resolved", capturedStatus)
	}
}

func TestReviewFraudAlert_InvalidStatus(t *testing.T) {
	svc := NewAdminService(&mockRepo{}, []byte("test-secret"))

	err := svc.ReviewFraudAlert(context.Background(), "alert-1", "invalid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReviewKYCSubmission_Approved(t *testing.T) {
	capturedReason := ""
	svc := NewAdminService(&mockRepo{
		reviewKYCSubmissionFn: func(ctx context.Context, id, status, reason string) error {
			capturedReason = reason
			return nil
		},
	}, []byte("test-secret"))

	err := svc.ReviewKYCSubmission(context.Background(), "kyc-1", "approved", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedReason != "" {
		t.Errorf("reason = %q, want empty string", capturedReason)
	}
}

func TestReviewKYCSubmission_Rejected(t *testing.T) {
	capturedReason := ""
	svc := NewAdminService(&mockRepo{
		reviewKYCSubmissionFn: func(ctx context.Context, id, status, reason string) error {
			capturedReason = reason
			return nil
		},
	}, []byte("test-secret"))

	err := svc.ReviewKYCSubmission(context.Background(), "kyc-1", "rejected", "bad document")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedReason != "bad document" {
		t.Errorf("reason = %q, want bad document", capturedReason)
	}
}

func TestReviewKYCSubmission_InvalidStatus(t *testing.T) {
	svc := NewAdminService(&mockRepo{}, []byte("test-secret"))

	err := svc.ReviewKYCSubmission(context.Background(), "kyc-1", "invalid", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReviewKYCSubmission_RejectedNoReason(t *testing.T) {
	svc := NewAdminService(&mockRepo{}, []byte("test-secret"))

	err := svc.ReviewKYCSubmission(context.Background(), "kyc-1", "rejected", "")
	if err == nil {
		t.Fatal("expected error for missing reason, got nil")
	}
}

func TestUpdateUserStatus(t *testing.T) {
	capturedID := ""
	capturedStatus := ""
	svc := NewAdminService(&mockRepo{
		updateUserStatusFn: func(ctx context.Context, id, kycStatus string) error {
			capturedID = id
			capturedStatus = kycStatus
			return nil
		},
	}, []byte("test-secret"))

	err := svc.UpdateUserStatus(context.Background(), "user-1", "approved")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != "user-1" {
		t.Errorf("id = %q, want user-1", capturedID)
	}
	if capturedStatus != "approved" {
		t.Errorf("status = %q, want approved", capturedStatus)
	}
}

func TestListMerchants_Pagination(t *testing.T) {
	capturedOffset := -1
	svc := NewAdminService(&mockRepo{
		listMerchantsFn: func(ctx context.Context, offset, limit int) ([]repository.Merchant, int, error) {
			capturedOffset = offset
			return nil, 0, nil
		},
	}, []byte("test-secret"))

	svc.ListMerchants(context.Background(), 2, 30)
	if capturedOffset != 30 {
		t.Errorf("offset = %d, want 30", capturedOffset)
	}
}

func TestGetMerchant_Success(t *testing.T) {
	svc := NewAdminService(&mockRepo{
		getMerchantFn: func(ctx context.Context, id string) (*repository.Merchant, error) {
			return &repository.Merchant{
				ID:      id,
				Name:    "Test Merchant",
				Email:   "merchant@test.com",
				Status:  "active",
				Country: "US",
			}, nil
		},
	}, []byte("test-secret"))

	m, err := svc.GetMerchant(context.Background(), "m-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "Test Merchant" {
		t.Errorf("Name = %q, want Test Merchant", m.Name)
	}
}

func TestListTransactions_Pagination(t *testing.T) {
	capturedOffset := -1
	svc := NewAdminService(&mockRepo{
		listTransactionsFn: func(ctx context.Context, offset, limit int) ([]repository.Transaction, int, error) {
			capturedOffset = offset
			return nil, 0, nil
		},
	}, []byte("test-secret"))

	svc.ListTransactions(context.Background(), 1, 10)
	if capturedOffset != 0 {
		t.Errorf("offset = %d, want 0", capturedOffset)
	}
}

func TestGetTransaction_Success(t *testing.T) {
	svc := NewAdminService(&mockRepo{
		getTransactionFn: func(ctx context.Context, id string) (*repository.Transaction, error) {
			return &repository.Transaction{
				ID:     id,
				Amount: 99.99,
				Status: "completed",
			}, nil
		},
	}, []byte("test-secret"))

	tx, err := svc.GetTransaction(context.Background(), "tx-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Amount != 99.99 {
		t.Errorf("Amount = %f, want 99.99", tx.Amount)
	}
}

func TestListFraudAlerts_Pagination(t *testing.T) {
	capturedOffset := -1
	svc := NewAdminService(&mockRepo{
		listFraudAlertsFn: func(ctx context.Context, offset, limit int) ([]repository.FraudAlert, int, error) {
			capturedOffset = offset
			return nil, 0, nil
		},
	}, []byte("test-secret"))

	svc.ListFraudAlerts(context.Background(), 3, 15)
	if capturedOffset != 30 {
		t.Errorf("offset = %d, want 30", capturedOffset)
	}
}

func TestReviewFraudAlert_Statuses(t *testing.T) {
	validStatuses := []string{"resolved", "false_positive", "investigated"}
	svc := NewAdminService(&mockRepo{
		reviewFraudAlertFn: func(ctx context.Context, id, status string) error {
			return nil
		},
	}, []byte("test-secret"))

	for _, s := range validStatuses {
		err := svc.ReviewFraudAlert(context.Background(), "a-1", s)
		if err != nil {
			t.Errorf("ReviewFraudAlert(%q): unexpected error: %v", s, err)
		}
	}
}

func TestReviewFraudAlert_RepoError(t *testing.T) {
	svc := NewAdminService(&mockRepo{
		reviewFraudAlertFn: func(ctx context.Context, id, status string) error {
			return errors.New("db error")
		},
	}, []byte("test-secret"))

	err := svc.ReviewFraudAlert(context.Background(), "a-1", "resolved")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
