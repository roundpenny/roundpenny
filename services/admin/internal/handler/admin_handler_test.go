// Copyright (c) 2026 RoundPenny. All rights reserved.

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/roundup-platform/services/admin/internal/repository"
	"github.com/roundup-platform/services/admin/internal/service"
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

var jwtSecret = []byte("handler-test-secret")

func adminToken(t *testing.T) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub":   "admin",
		"email": "admin@roundpenny.com",
		"role":  "admin",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
		"jti":   "test-jti",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return s
}

func nonAdminToken(t *testing.T) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub":   "user-1",
		"email": "user@test.com",
		"role":  "user",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
		"jti":   "test-jti",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return s
}

func newHandler(mock *mockRepo) *AdminHandler {
	svc := service.NewAdminService(mock, jwtSecret)
	return NewAdminHandler(svc)
}

func readBody(t *testing.T, rec *httptest.ResponseRecorder, dest interface{}) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(dest); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	h := newHandler(&mockRepo{})
	handler := h.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
	var resp map[string]string
	readBody(t, rec, &resp)
	if resp["error"] != "unauthorized" {
		t.Errorf("error = %q, want unauthorized", resp["error"])
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	h := newHandler(&mockRepo{})
	handler := h.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuthMiddleware_NoBearerPrefix(t *testing.T) {
	h := newHandler(&mockRepo{})
	handler := h.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuthMiddleware_Forbidden(t *testing.T) {
	h := newHandler(&mockRepo{})
	handler := h.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+nonAdminToken(t))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
	var resp map[string]string
	readBody(t, rec, &resp)
	if resp["error"] != "forbidden" {
		t.Errorf("error = %q, want forbidden", resp["error"])
	}
}

func TestAuthMiddleware_Success(t *testing.T) {
	called := false
	h := newHandler(&mockRepo{})
	handler := h.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !called {
		t.Error("next handler was not called")
	}
}

func TestDashboardHandler(t *testing.T) {
	h := newHandler(&mockRepo{
		getStatsFn: func(ctx context.Context) (*repository.AdminStats, error) {
			return &repository.AdminStats{
				TotalUsers:        100,
				PendingKYC:        5,
				ActiveMerchants:   20,
				TotalTransactions: 5000,
				OpenFraudAlerts:   2,
				PendingPayments:   10,
			}, nil
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.GetStats)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	readBody(t, rec, &resp)
	if resp["total_users"].(float64) != 100 {
		t.Errorf("total_users = %v, want 100", resp["total_users"])
	}
	if resp["pending_kyc"].(float64) != 5 {
		t.Errorf("pending_kyc = %v, want 5", resp["pending_kyc"])
	}
}

func TestDashboardHandler_Error(t *testing.T) {
	h := newHandler(&mockRepo{
		getStatsFn: func(ctx context.Context) (*repository.AdminStats, error) {
			return nil, http.ErrAbortHandler
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.GetStats)(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestLoginHandler_Success(t *testing.T) {
	h := newHandler(&mockRepo{})

	body := `{"email":"admin@roundpenny.com","password":"admin123"}`
	req := httptest.NewRequest("POST", "/v1/admin/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	readBody(t, rec, &resp)
	if resp["access_token"] == "" {
		t.Error("expected access_token")
	}
	if resp["token_type"] != "Bearer" {
		t.Errorf("token_type = %v, want Bearer", resp["token_type"])
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	h := newHandler(&mockRepo{})

	body := `{"email":"wrong@email.com","password":"wrong"}`
	req := httptest.NewRequest("POST", "/v1/admin/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
	var resp map[string]string
	readBody(t, rec, &resp)
	if resp["error"] != "invalid credentials" {
		t.Errorf("error = %q, want invalid credentials", resp["error"])
	}
}

func TestLoginHandler_BadRequest(t *testing.T) {
	h := newHandler(&mockRepo{})

	req := httptest.NewRequest("POST", "/v1/admin/login", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestGetUserHandler(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	h := newHandler(&mockRepo{
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
				UpdatedAt:     now,
			}, nil
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/users/user-1", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.SetPathValue("id", "user-1")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.GetUser)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	readBody(t, rec, &resp)
	if resp["id"] != "user-1" {
		t.Errorf("id = %v, want user-1", resp["id"])
	}
	if resp["email"] != "alice@test.com" {
		t.Errorf("email = %v, want alice@test.com", resp["email"])
	}
}

func TestGetUserHandler_NotFound(t *testing.T) {
	h := newHandler(&mockRepo{
		getUserFn: func(ctx context.Context, id string) (*repository.User, error) {
			return nil, nil
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/users/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.GetUser)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	var resp map[string]string
	readBody(t, rec, &resp)
	if resp["error"] != "user not found" {
		t.Errorf("error = %q, want user not found", resp["error"])
	}
}

func TestGetUserHandler_Error(t *testing.T) {
	h := newHandler(&mockRepo{
		getUserFn: func(ctx context.Context, id string) (*repository.User, error) {
			return nil, http.ErrAbortHandler
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/users/user-1", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.SetPathValue("id", "user-1")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.GetUser)(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestListKycHandler(t *testing.T) {
	now := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	h := newHandler(&mockRepo{
		listKYCSubmissionsFn: func(ctx context.Context, offset, limit int) ([]repository.KYCSubmission, int, error) {
			return []repository.KYCSubmission{
				{ID: "kyc-1", UserID: "user-1", FullName: "Bob", DocumentType: "passport", Status: "pending", SubmittedAt: now},
			}, 1, nil
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/kyc-submissions?page=1&page_size=20", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.ListKYCSubmissions)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	readBody(t, rec, &resp)
	if resp["total"].(float64) != 1 {
		t.Errorf("total = %v, want 1", resp["total"])
	}
}

func TestListKycHandler_Error(t *testing.T) {
	h := newHandler(&mockRepo{
		listKYCSubmissionsFn: func(ctx context.Context, offset, limit int) ([]repository.KYCSubmission, int, error) {
			return nil, 0, http.ErrAbortHandler
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/kyc-submissions", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.ListKYCSubmissions)(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestListUsersHandler(t *testing.T) {
	now := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	h := newHandler(&mockRepo{
		listUsersFn: func(ctx context.Context, offset, limit int) ([]repository.User, int, error) {
			return []repository.User{
				{ID: "user-1", Email: "a@test.com", FullName: "Alice", Role: "user", CreatedAt: now, UpdatedAt: now},
			}, 1, nil
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/users?page=1&page_size=20", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.ListUsers)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var resp map[string]interface{}
	readBody(t, rec, &resp)
	if resp["total"].(float64) != 1 {
		t.Errorf("total = %v, want 1", resp["total"])
	}
}

func TestListUsersHandler_Error(t *testing.T) {
	h := newHandler(&mockRepo{
		listUsersFn: func(ctx context.Context, offset, limit int) ([]repository.User, int, error) {
			return nil, 0, http.ErrAbortHandler
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.ListUsers)(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestUpdateUserStatusHandler(t *testing.T) {
	h := newHandler(&mockRepo{
		updateUserStatusFn: func(ctx context.Context, id, kycStatus string) error {
			return nil
		},
	})

	body := `{"kyc_status":"approved"}`
	req := httptest.NewRequest("PUT", "/v1/admin/users/user-1/status", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "user-1")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.UpdateUserStatus)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var resp map[string]string
	readBody(t, rec, &resp)
	if resp["status"] != "updated" {
		t.Errorf("status = %q, want updated", resp["status"])
	}
}

func TestUpdateUserStatusHandler_BadRequest(t *testing.T) {
	h := newHandler(&mockRepo{})

	req := httptest.NewRequest("PUT", "/v1/admin/users/user-1/status", strings.NewReader("not json"))
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "user-1")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.UpdateUserStatus)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestListMerchantsHandler(t *testing.T) {
	h := newHandler(&mockRepo{
		listMerchantsFn: func(ctx context.Context, offset, limit int) ([]repository.Merchant, int, error) {
			return []repository.Merchant{
				{ID: "m-1", Name: "Test Merchant", Status: "active"},
			}, 1, nil
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/merchants", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.ListMerchants)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestGetMerchantHandler(t *testing.T) {
	h := newHandler(&mockRepo{
		getMerchantFn: func(ctx context.Context, id string) (*repository.Merchant, error) {
			return &repository.Merchant{ID: id, Name: "Test", Status: "active"}, nil
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/merchants/m-1", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.SetPathValue("id", "m-1")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.GetMerchant)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestGetMerchantHandler_NotFound(t *testing.T) {
	h := newHandler(&mockRepo{
		getMerchantFn: func(ctx context.Context, id string) (*repository.Merchant, error) {
			return nil, nil
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/merchants/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.GetMerchant)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestListTransactionsHandler(t *testing.T) {
	h := newHandler(&mockRepo{
		listTransactionsFn: func(ctx context.Context, offset, limit int) ([]repository.Transaction, int, error) {
			return []repository.Transaction{
				{ID: "tx-1", Amount: 99.99, Status: "completed"},
			}, 1, nil
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.ListTransactions)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestGetTransactionHandler(t *testing.T) {
	h := newHandler(&mockRepo{
		getTransactionFn: func(ctx context.Context, id string) (*repository.Transaction, error) {
			return &repository.Transaction{ID: id, Amount: 49.99, Status: "pending"}, nil
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/transactions/tx-1", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.SetPathValue("id", "tx-1")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.GetTransaction)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestGetTransactionHandler_NotFound(t *testing.T) {
	h := newHandler(&mockRepo{
		getTransactionFn: func(ctx context.Context, id string) (*repository.Transaction, error) {
			return nil, nil
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/transactions/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.GetTransaction)(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestListFraudAlertsHandler(t *testing.T) {
	h := newHandler(&mockRepo{
		listFraudAlertsFn: func(ctx context.Context, offset, limit int) ([]repository.FraudAlert, int, error) {
			return []repository.FraudAlert{
				{ID: "fa-1", Severity: "high", Status: "open"},
			}, 1, nil
		},
	})

	req := httptest.NewRequest("GET", "/v1/admin/fraud-alerts", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.ListFraudAlerts)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestReviewFraudAlertHandler(t *testing.T) {
	h := newHandler(&mockRepo{
		reviewFraudAlertFn: func(ctx context.Context, id, status string) error {
			return nil
		},
	})

	body := `{"status":"resolved"}`
	req := httptest.NewRequest("POST", "/v1/admin/fraud-alerts/fa-1/review", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "fa-1")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.ReviewFraudAlert)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestReviewFraudAlertHandler_Invalid(t *testing.T) {
	h := newHandler(&mockRepo{
		reviewFraudAlertFn: func(ctx context.Context, id, status string) error {
			return service.ErrInvalidCredentials
		},
	})

	body := `{"status":"invalid"}`
	req := httptest.NewRequest("POST", "/v1/admin/fraud-alerts/fa-1/review", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "fa-1")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.ReviewFraudAlert)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestReviewFraudAlertHandler_BadRequest(t *testing.T) {
	h := newHandler(&mockRepo{})

	req := httptest.NewRequest("POST", "/v1/admin/fraud-alerts/fa-1/review", strings.NewReader("not json"))
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "fa-1")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.ReviewFraudAlert)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestReviewKycSubmissionHandler(t *testing.T) {
	h := newHandler(&mockRepo{
		reviewKYCSubmissionFn: func(ctx context.Context, id, status, reason string) error {
			return nil
		},
	})

	body := `{"status":"approved","rejection_reason":""}`
	req := httptest.NewRequest("POST", "/v1/admin/kyc-submissions/kyc-1/review", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "kyc-1")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.ReviewKYCSubmission)(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestReviewKycSubmissionHandler_Invalid(t *testing.T) {
	h := newHandler(&mockRepo{
		reviewKYCSubmissionFn: func(ctx context.Context, id, status, reason string) error {
			return service.ErrInvalidCredentials
		},
	})

	body := `{"status":"invalid","rejection_reason":""}`
	req := httptest.NewRequest("POST", "/v1/admin/kyc-submissions/kyc-1/review", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "kyc-1")
	rec := httptest.NewRecorder()
	h.AuthMiddleware(h.ReviewKYCSubmission)(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestLogoutHandler(t *testing.T) {
	h := newHandler(&mockRepo{})

	req := httptest.NewRequest("POST", "/v1/admin/logout", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t))
	rec := httptest.NewRecorder()
	h.Logout(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rec.Code)
	}
}
