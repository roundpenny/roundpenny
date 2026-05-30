package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/roundup-platform/pkg/cache"
	"github.com/roundup-platform/pkg/crypto"
	"github.com/roundup-platform/pkg/email"
	"github.com/roundup-platform/pkg/kyc"
	"github.com/roundup-platform/services/auth/internal/repository"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *repository.User) error
	GetUserByEmail(ctx context.Context, email string) (*repository.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*repository.User, error)
	SaveRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, tokenHash string) (bool, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeUserTokens(ctx context.Context, userID uuid.UUID) error
	SaveMFASecret(ctx context.Context, userID uuid.UUID, secret string) error
	EnableMFA(ctx context.Context, userID uuid.UUID, backupCodes []string) error
	DisableMFA(ctx context.Context, userID uuid.UUID) error
	IncrementFailedLogins(ctx context.Context, userID uuid.UUID) error
	ResetFailedLogins(ctx context.Context, userID uuid.UUID) error
	LockUser(ctx context.Context, userID uuid.UUID, until time.Time) error
	UpsertOAuthAccount(ctx context.Context, userID uuid.UUID, provider, providerUserID, accessToken, refreshToken string, expiresAt *time.Time) error
	GetOAuthAccount(ctx context.Context, provider, providerUserID string) (*repository.OAuthAccount, error)
	CreateEmailVerification(ctx context.Context, userID uuid.UUID, email, token string, expiresAt time.Time) error
	GetEmailVerificationByToken(ctx context.Context, token string) (*repository.EmailVerification, error)
	VerifyEmail(ctx context.Context, token string) error
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
	CreateKYCSubmission(ctx context.Context, sub *repository.KYCSubmission) error
	GetKYCSubmissionByUserID(ctx context.Context, userID uuid.UUID) (*repository.KYCSubmission, error)
	UpdateKYCStatus(ctx context.Context, userID uuid.UUID, status string) error
}

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenRevoked       = errors.New("token has been revoked")
	ErrMFARequired        = errors.New("MFA verification required")
	ErrInvalidMFA         = errors.New("invalid MFA code")
	ErrAccountLocked      = errors.New("account is temporarily locked")
	ErrMFANotEnabled      = errors.New("MFA is not enabled")
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
)

type AuthService struct {
	repo        AuthRepository
	jwtSecret   []byte
	emailClient *email.Client
	kycClient   *kyc.Client
	cacheClient *cache.Client
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(repo AuthRepository, jwtSecret string) *AuthService {
	return &AuthService{
		repo:        repo,
		jwtSecret:   []byte(jwtSecret),
		emailClient: email.NewClient(),
		kycClient:   kyc.NewClient(),
		cacheClient: cache.NewClient(),
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"full_name"`
	Phone    string `json:"phone,omitempty"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	MFARequired  bool         `json:"mfa_required,omitempty"`
	MFAToken     string       `json:"mfa_token,omitempty"`
	User         *UserResponse `json:"user"`
}

type UserResponse struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	FullName      string    `json:"full_name"`
	EmailVerified bool      `json:"email_verified"`
	KYCStatus     string    `json:"kyc_status"`
	Role          string    `json:"role"`
	MFAEnabled    bool      `json:"mfa_enabled"`
	CreatedAt     time.Time `json:"created_at"`
}

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 7 * 24 * time.Hour
	maxFailedLogins      = 5
	lockoutDuration      = 15 * time.Minute
)

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &repository.User{
		Email:        req.Email,
		PasswordHash: hash,
		FullName:     req.Name,
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	s.emailClient.Send(email.SendEmailParams{
		To:      user.Email,
		Subject: "Welcome to RoundPenny!",
		HTML:    "<h1>Welcome to RoundPenny!</h1><p>Start saving automatically with every purchase. Your spare change adds up fast.</p>",
	})

	return s.generateTokens(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := crypto.ComparePassword(user.PasswordHash, password); err != nil {
		_ = s.repo.IncrementFailedLogins(ctx, user.ID)
		return nil, ErrInvalidCredentials
	}

	_ = s.repo.ResetFailedLogins(ctx, user.ID)

	if user.MFAEnabled {
		mfaToken, err := s.createMFASession(ctx, user.ID)
		if err != nil {
			return nil, fmt.Errorf("create MFA session: %w", err)
		}
		return &AuthResponse{
			MFARequired: true,
			MFAToken:    mfaToken,
			User:        toUserResponse(user),
		}, nil
	}

	return s.generateTokens(ctx, user)
}

func (s *AuthService) VerifyMFA(ctx context.Context, mfaToken, code string) (*AuthResponse, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(mfaToken, claims, func(t *jwt.Token) (any, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrTokenExpired
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !user.MFAEnabled || user.MFASecret == nil {
		return nil, ErrMFANotEnabled
	}

	valid := totp.Validate(code, *user.MFASecret)
	if !valid {
		return nil, ErrInvalidMFA
	}

	return s.generateTokens(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (any, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrTokenExpired
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	blacklisted, _ := s.cacheClient.Exists(ctx, "logout:"+userID.String())
	if blacklisted {
		return nil, ErrTokenRevoked
	}

	tokenHash := hashToken(refreshToken)
	valid, err := s.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("check refresh token: %w", err)
	}
	if !valid {
		return nil, ErrTokenRevoked
	}

	if err := s.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("revoke old token: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return s.generateTokens(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	s.cacheClient.Set(ctx, "logout:"+userID.String(), "1", refreshTokenDuration)
	return s.repo.RevokeUserTokens(ctx, userID)
}

func (s *AuthService) SetupMFA(ctx context.Context, userID uuid.UUID) (string, string, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", ErrUserNotFound
	}

	if user.MFAEnabled {
		return "", "", errors.New("MFA already enabled")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "RoundupPlatform",
		AccountName: user.Email,
	})
	if err != nil {
		return "", "", fmt.Errorf("generate TOTP key: %w", err)
	}

	if err := s.repo.SaveMFASecret(ctx, userID, key.Secret()); err != nil {
		return "", "", fmt.Errorf("save MFA secret: %w", err)
	}

	return key.Secret(), key.URL(), nil
}

func (s *AuthService) EnableMFA(ctx context.Context, userID uuid.UUID, code string) ([]string, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if user.MFASecret == nil {
		return nil, errors.New("MFA not set up")
	}

	if !totp.Validate(code, *user.MFASecret) {
		return nil, ErrInvalidMFA
	}

	backupCodes, err := repository.GenerateBackupCodes()
	if err != nil {
		return nil, fmt.Errorf("generate backup codes: %w", err)
	}

	if err := s.repo.EnableMFA(ctx, userID, backupCodes); err != nil {
		return nil, fmt.Errorf("enable MFA: %w", err)
	}

	return backupCodes, nil
}

func (s *AuthService) DisableMFA(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DisableMFA(ctx, userID)
}

func (s *AuthService) OAuthLogin(ctx context.Context, provider, code, state, redirectURI string) (*AuthResponse, error) {
	providerUserID := fmt.Sprintf("%s_%x", provider, sha256.Sum256([]byte(code)))

	account, err := s.repo.GetOAuthAccount(ctx, provider, providerUserID)
	if err == nil {
		user, err := s.repo.GetUserByID(ctx, account.UserID)
		if err != nil {
			return nil, ErrUserNotFound
		}
		return s.generateTokens(ctx, user)
	}

	email := fmt.Sprintf("%s-%s@oauth.%s.com", provider, providerUserID[:8], provider)
	placeholderHash, _ := crypto.HashPassword(uuid.New().String())
	user := &repository.User{
		Email:        email,
		PasswordHash: placeholderHash,
		FullName:     fmt.Sprintf("%s User", provider),
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("create oauth user: %w", err)
	}
	if err := s.repo.UpsertOAuthAccount(ctx, user.ID, provider, providerUserID, "", "", nil); err != nil {
		return nil, fmt.Errorf("save oauth account: %w", err)
	}

	return s.generateTokens(ctx, user)
}

func (s *AuthService) SendEmailVerification(ctx context.Context, userID uuid.UUID) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}
	if user.EmailVerified {
		return errors.New("email already verified")
	}

	token, err := GenerateRandomString(32)
	if err != nil {
		return fmt.Errorf("generate token: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.repo.CreateEmailVerification(ctx, userID, user.Email, token, expiresAt); err != nil {
		return fmt.Errorf("create verification: %w", err)
	}

	verifyLink := fmt.Sprintf("https://roundpenny.com/verify-email?token=%s", token)
	s.emailClient.Send(email.SendEmailParams{
		To:      user.Email,
		Subject: "Verify your email address",
		HTML:    fmt.Sprintf("<h1>Verify your email</h1><p>Click <a href='%s'>here</a> to verify your email address. This link expires in 24 hours.</p>", verifyLink),
	})

	return nil
}

func (s *AuthService) ConfirmEmailVerification(ctx context.Context, token string) error {
	verification, err := s.repo.GetEmailVerificationByToken(ctx, token)
	if err != nil {
		return ErrInvalidVerificationToken
	}
	if verification.VerifiedAt != nil {
		return errors.New("email already verified")
	}
	if time.Now().After(verification.ExpiresAt) {
		return ErrInvalidVerificationToken
	}
	if err := s.repo.VerifyEmail(ctx, token); err != nil {
		return fmt.Errorf("verify email: %w", err)
	}
	if err := s.repo.MarkEmailVerified(ctx, verification.UserID); err != nil {
		return fmt.Errorf("mark verified: %w", err)
	}
	return nil
}

func (s *AuthService) SubmitKYC(ctx context.Context, userID uuid.UUID, fullName, documentType, documentNumber string) (*repository.KYCSubmission, error) {
	sub := &repository.KYCSubmission{
		UserID:         userID,
		FullName:       fullName,
		DocumentType:   documentType,
		DocumentNumber: documentNumber,
	}
	if err := s.repo.CreateKYCSubmission(ctx, sub); err != nil {
		return nil, fmt.Errorf("create kyc submission: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	parts := strings.SplitN(fullName, " ", 2)
	firstName := parts[0]
	lastName := ""
	if len(parts) > 1 {
		lastName = parts[1]
	}

	applicant, err := s.kycClient.CreateApplicant(firstName, lastName, user.Email)
	if err != nil {
		slog.Error("kyc create applicant failed", "error", err)
		s.repo.UpdateKYCStatus(ctx, userID, "pending")
		sub.Status = "pending"
		return sub, nil
	}

	check, err := s.kycClient.CreateCheck(applicant.ID, nil)
	if err != nil {
		slog.Error("kyc create check failed", "error", err)
		s.repo.UpdateKYCStatus(ctx, userID, "pending")
		sub.Status = "pending"
		return sub, nil
	}

	status := "pending"
	if check.Status == "complete" && check.Result == "clear" {
		status = "approved"
	} else if check.Result == "consider" || check.Result == "unidentified" {
		status = "rejected"
	}

	if err := s.repo.UpdateKYCStatus(ctx, userID, status); err != nil {
		return nil, fmt.Errorf("update kyc status: %w", err)
	}
	sub.Status = status

	slog.Info("kyc check completed",
		"user", userID,
		"status", status,
		"onfido_check", check.ID,
	)

	return sub, nil
}

func (s *AuthService) GetKYCStatus(ctx context.Context, userID uuid.UUID) (*repository.KYCSubmission, error) {
	sub, err := s.repo.GetKYCSubmissionByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get kyc status: %w", err)
	}
	return sub, nil
}

func (s *AuthService) GetUser(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return toUserResponse(user), nil
}

func (s *AuthService) Authenticate(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (s *AuthService) createMFASession(ctx context.Context, userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID.String(),
		Role:   "mfa_pending",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (s *AuthService) generateTokens(ctx context.Context, user *repository.User) (*AuthResponse, error) {
	now := time.Now()
	accessExpiry := now.Add(accessTokenDuration)
	refreshExpiry := now.Add(refreshTokenDuration)

	accessID, _ := uuid.NewRandom()
	refreshID, _ := uuid.NewRandom()

	accessClaims := &Claims{
		UserID: user.ID.String(),
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessID.String(),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := &Claims{
		UserID: user.ID.String(),
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshID.String(),
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID.String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	refreshHash := hashToken(refreshStr)
	if err := s.repo.SaveRefreshToken(ctx, user.ID, refreshHash, refreshExpiry); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresIn:    900,
		User:         toUserResponse(user),
	}, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func toUserResponse(user *repository.User) *UserResponse {
	return &UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		FullName:      user.FullName,
		EmailVerified: user.EmailVerified,
		KYCStatus:     user.KYCStatus,
		Role:          user.Role,
		MFAEnabled:    user.MFASecret != nil,
		CreatedAt:     user.CreatedAt,
	}
}

func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateTOTPSecret() (string, error) {
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(secret), "="), nil
}
