package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/crypto"
	"github.com/roundup-platform/services/auth/internal/repository"
)

type mockAuthRepo struct {
	createUserFn          func(ctx context.Context, user *repository.User) error
	getUserByEmailFn      func(ctx context.Context, email string) (*repository.User, error)
	getUserByIDFn         func(ctx context.Context, id uuid.UUID) (*repository.User, error)
	saveRefreshTokenFn    func(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	getRefreshTokenFn     func(ctx context.Context, tokenHash string) (bool, error)
	revokeRefreshTokenFn  func(ctx context.Context, tokenHash string) error
	revokeUserTokensFn    func(ctx context.Context, userID uuid.UUID) error
	saveMFASecretFn       func(ctx context.Context, userID uuid.UUID, secret string) error
	enableMFAFn           func(ctx context.Context, userID uuid.UUID, backupCodes []string) error
	disableMFAFn          func(ctx context.Context, userID uuid.UUID) error
	incrementFailedLogins       func(ctx context.Context, userID uuid.UUID) error
	resetFailedLogins           func(ctx context.Context, userID uuid.UUID) error
	lockUserFn                  func(ctx context.Context, userID uuid.UUID, until time.Time) error
	upsertOAuthAccountFn        func(ctx context.Context, userID uuid.UUID, provider, providerUserID, accessToken, refreshToken string, expiresAt *time.Time) error
	getOAuthAccountFn           func(ctx context.Context, provider, providerUserID string) (*repository.OAuthAccount, error)
	createEmailVerificationFn   func(ctx context.Context, userID uuid.UUID, email, token string, expiresAt time.Time) error
	getEmailVerificationByToken func(ctx context.Context, token string) (*repository.EmailVerification, error)
	verifyEmailFn               func(ctx context.Context, token string) error
	markEmailVerifiedFn         func(ctx context.Context, userID uuid.UUID) error
	createKYCSubmissionFn       func(ctx context.Context, sub *repository.KYCSubmission) error
	getKYCSubmissionByUserIDFn  func(ctx context.Context, userID uuid.UUID) (*repository.KYCSubmission, error)
	updateKYCStatusFn           func(ctx context.Context, userID uuid.UUID, status string) error
}

func (m *mockAuthRepo) CreateUser(ctx context.Context, user *repository.User) error {
	return m.createUserFn(ctx, user)
}
func (m *mockAuthRepo) GetUserByEmail(ctx context.Context, email string) (*repository.User, error) {
	return m.getUserByEmailFn(ctx, email)
}
func (m *mockAuthRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*repository.User, error) {
	return m.getUserByIDFn(ctx, id)
}
func (m *mockAuthRepo) SaveRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return m.saveRefreshTokenFn(ctx, userID, tokenHash, expiresAt)
}
func (m *mockAuthRepo) GetRefreshToken(ctx context.Context, tokenHash string) (bool, error) {
	return m.getRefreshTokenFn(ctx, tokenHash)
}
func (m *mockAuthRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	return m.revokeRefreshTokenFn(ctx, tokenHash)
}
func (m *mockAuthRepo) RevokeUserTokens(ctx context.Context, userID uuid.UUID) error {
	return m.revokeUserTokensFn(ctx, userID)
}
func (m *mockAuthRepo) SaveMFASecret(ctx context.Context, userID uuid.UUID, secret string) error {
	return m.saveMFASecretFn(ctx, userID, secret)
}
func (m *mockAuthRepo) EnableMFA(ctx context.Context, userID uuid.UUID, backupCodes []string) error {
	return m.enableMFAFn(ctx, userID, backupCodes)
}
func (m *mockAuthRepo) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	return m.disableMFAFn(ctx, userID)
}
func (m *mockAuthRepo) IncrementFailedLogins(ctx context.Context, userID uuid.UUID) error {
	return m.incrementFailedLogins(ctx, userID)
}
func (m *mockAuthRepo) ResetFailedLogins(ctx context.Context, userID uuid.UUID) error {
	return m.resetFailedLogins(ctx, userID)
}
func (m *mockAuthRepo) LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error {
	return m.lockUserFn(ctx, userID, until)
}
func (m *mockAuthRepo) UpsertOAuthAccount(ctx context.Context, userID uuid.UUID, provider, providerUserID, accessToken, refreshToken string, expiresAt *time.Time) error {
	if m.upsertOAuthAccountFn != nil {
		return m.upsertOAuthAccountFn(ctx, userID, provider, providerUserID, accessToken, refreshToken, expiresAt)
	}
	return nil
}
func (m *mockAuthRepo) GetOAuthAccount(ctx context.Context, provider, providerUserID string) (*repository.OAuthAccount, error) {
	if m.getOAuthAccountFn != nil {
		return m.getOAuthAccountFn(ctx, provider, providerUserID)
	}
	return nil, errors.New("not found")
}
func (m *mockAuthRepo) CreateEmailVerification(ctx context.Context, userID uuid.UUID, email, token string, expiresAt time.Time) error {
	if m.createEmailVerificationFn != nil {
		return m.createEmailVerificationFn(ctx, userID, email, token, expiresAt)
	}
	return nil
}
func (m *mockAuthRepo) GetEmailVerificationByToken(ctx context.Context, token string) (*repository.EmailVerification, error) {
	if m.getEmailVerificationByToken != nil {
		return m.getEmailVerificationByToken(ctx, token)
	}
	return nil, errors.New("not found")
}
func (m *mockAuthRepo) VerifyEmail(ctx context.Context, token string) error {
	if m.verifyEmailFn != nil {
		return m.verifyEmailFn(ctx, token)
	}
	return nil
}
func (m *mockAuthRepo) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	if m.markEmailVerifiedFn != nil {
		return m.markEmailVerifiedFn(ctx, userID)
	}
	return nil
}
func (m *mockAuthRepo) CreateKYCSubmission(ctx context.Context, sub *repository.KYCSubmission) error {
	if m.createKYCSubmissionFn != nil {
		return m.createKYCSubmissionFn(ctx, sub)
	}
	return nil
}
func (m *mockAuthRepo) GetKYCSubmissionByUserID(ctx context.Context, userID uuid.UUID) (*repository.KYCSubmission, error) {
	if m.getKYCSubmissionByUserIDFn != nil {
		return m.getKYCSubmissionByUserIDFn(ctx, userID)
	}
	return nil, errors.New("not found")
}
func (m *mockAuthRepo) UpdateKYCStatus(ctx context.Context, userID uuid.UUID, status string) error {
	if m.updateKYCStatusFn != nil {
		return m.updateKYCStatusFn(ctx, userID, status)
	}
	return nil
}

func TestRegister_Success(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	svc := NewAuthService(&mockAuthRepo{
		createUserFn: func(ctx context.Context, user *repository.User) error {
			user.ID = userID
			user.CreatedAt = now
			user.EmailVerified = false
			user.KYCStatus = "pending"
			user.Role = "user"
			return nil
		},
		saveRefreshTokenFn: func(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
			return nil
		},
	}, "test-secret")

	resp, err := svc.Register(context.Background(), RegisterRequest{Email: "test@test.com", Password: "password123", Name: "Test User"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.User.ID != userID {
		t.Fatalf("expected user ID %v, got %v", userID, resp.User.ID)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected access token")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := NewAuthService(&mockAuthRepo{
		createUserFn: func(ctx context.Context, user *repository.User) error {
			return errors.New("duplicate key value")
		},
	}, "test-secret")

	_, err := svc.Register(context.Background(), RegisterRequest{Email: "dup@test.com", Password: "password123", Name: "Dup"})
	if err == nil {
		t.Fatal("expected error on duplicate")
	}
}

func TestLogin_Success(t *testing.T) {
	hash, _ := crypto.HashPassword("correct-password")
	svc := NewAuthService(&mockAuthRepo{
		getUserByEmailFn: func(ctx context.Context, email string) (*repository.User, error) {
			return &repository.User{ID: uuid.New(), Email: email, PasswordHash: hash, FullName: "T", Role: "user", CreatedAt: time.Now()}, nil
		},
		saveRefreshTokenFn: func(ctx context.Context, uid uuid.UUID, th string, et time.Time) error { return nil },
		resetFailedLogins: func(ctx context.Context, uid uuid.UUID) error { return nil },
	}, "test-secret")

	resp, err := svc.Login(context.Background(), "t@t.com", "correct-password")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected token")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	hash, _ := crypto.HashPassword("correct-password")
	svc := NewAuthService(&mockAuthRepo{
		getUserByEmailFn: func(ctx context.Context, email string) (*repository.User, error) {
			return &repository.User{ID: uuid.New(), Email: email, PasswordHash: hash, FullName: "T", Role: "user", CreatedAt: time.Now()}, nil
		},
		incrementFailedLogins: func(ctx context.Context, uid uuid.UUID) error { return nil },
	}, "test-secret")

	_, err := svc.Login(context.Background(), "t@t.com", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_MFARequired(t *testing.T) {
	secret := "TESTSECRET"
	hash, _ := crypto.HashPassword("pwd")
	svc := NewAuthService(&mockAuthRepo{
		getUserByEmailFn: func(ctx context.Context, email string) (*repository.User, error) {
			return &repository.User{ID: uuid.New(), Email: email, PasswordHash: hash, FullName: "T", Role: "user", MFAEnabled: true, MFASecret: &secret, CreatedAt: time.Now()}, nil
		},
		resetFailedLogins: func(ctx context.Context, uid uuid.UUID) error { return nil },
	}, "test-secret")

	resp, err := svc.Login(context.Background(), "t@t.com", "pwd")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !resp.MFARequired {
		t.Fatal("expected MFARequired")
	}
	if resp.MFAToken == "" {
		t.Fatal("expected MFA token")
	}
}

func TestRefresh_Rotation(t *testing.T) {
	userID := uuid.New()
	hash, _ := crypto.HashPassword("pwd")
	revoked := false

	svc := NewAuthService(&mockAuthRepo{
		getUserByEmailFn: func(ctx context.Context, email string) (*repository.User, error) {
			return &repository.User{ID: userID, Email: email, PasswordHash: hash, FullName: "T", Role: "user", CreatedAt: time.Now()}, nil
		},
		saveRefreshTokenFn: func(ctx context.Context, uid uuid.UUID, th string, et time.Time) error { return nil },
		resetFailedLogins: func(ctx context.Context, uid uuid.UUID) error { return nil },
		getRefreshTokenFn: func(ctx context.Context, tokenHash string) (bool, error) { return !revoked, nil },
		revokeRefreshTokenFn: func(ctx context.Context, tokenHash string) error { revoked = true; return nil },
		getUserByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.User, error) {
			return &repository.User{ID: userID, FullName: "T", Role: "user", CreatedAt: time.Now()}, nil
		},
	}, "test-secret")

	loginResp, err := svc.Login(context.Background(), "t@t.com", "pwd")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	resp, err := svc.Refresh(context.Background(), loginResp.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("expected new access token")
	}
	if resp.RefreshToken == loginResp.RefreshToken {
		t.Fatal("expected new refresh token after rotation")
	}
}

func TestRefresh_RevokedToken(t *testing.T) {
	svc := NewAuthService(&mockAuthRepo{
		getRefreshTokenFn: func(ctx context.Context, tokenHash string) (bool, error) { return false, nil },
	}, "test-secret")

	_, err := svc.Refresh(context.Background(), "some-token-that-will-fail-jwt-parsing")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestSetupMFA_Success(t *testing.T) {
	userID := uuid.New()
	svc := NewAuthService(&mockAuthRepo{
		getUserByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.User, error) {
			return &repository.User{ID: userID, Email: "t@t.com", FullName: "T", Role: "user", CreatedAt: time.Now()}, nil
		},
		saveMFASecretFn: func(ctx context.Context, uid uuid.UUID, secret string) error { return nil },
	}, "test-secret")

	secret, url, err := svc.SetupMFA(context.Background(), userID)
	if err != nil {
		t.Fatalf("setup MFA: %v", err)
	}
	if secret == "" {
		t.Fatal("expected secret")
	}
	if url == "" {
		t.Fatal("expected URL")
	}
}

func TestGetUser_Success(t *testing.T) {
	userID := uuid.New()
	svc := NewAuthService(&mockAuthRepo{
		getUserByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.User, error) {
			return &repository.User{ID: id, Email: "t@t.com", FullName: "T", Role: "user", CreatedAt: time.Now()}, nil
		},
	}, "test-secret")

	user, err := svc.GetUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != userID {
		t.Fatalf("expected user ID %v, got %v", userID, user.ID)
	}
}

func TestLogout(t *testing.T) {
	revoked := false
	svc := NewAuthService(&mockAuthRepo{
		revokeUserTokensFn: func(ctx context.Context, userID uuid.UUID) error { revoked = true; return nil },
	}, "test-secret")

	if err := svc.Logout(context.Background(), uuid.New()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !revoked {
		t.Fatal("expected RevokeUserTokens to be called")
	}
}

func TestAuthenticate_ValidToken(t *testing.T) {
	svc := NewAuthService(&mockAuthRepo{
		createUserFn: func(ctx context.Context, user *repository.User) error {
			user.ID = uuid.New()
			user.CreatedAt = time.Now()
			user.EmailVerified = false
			user.KYCStatus = "pending"
			user.Role = "user"
			return nil
		},
		saveRefreshTokenFn: func(ctx context.Context, uid uuid.UUID, th string, et time.Time) error { return nil },
	}, "test-secret")

	resp, err := svc.Register(context.Background(), RegisterRequest{Email: "a@b.com", Password: "password123", Name: "A"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	claims, err := svc.Authenticate(resp.AccessToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if claims.UserID == "" {
		t.Fatal("expected UserID in claims")
	}
}

func TestAuthenticate_InvalidToken(t *testing.T) {
	svc := NewAuthService(&mockAuthRepo{}, "test-secret")
	_, err := svc.Authenticate("invalid-token")
	if err == nil {
		t.Fatal("expected error")
	}
}
