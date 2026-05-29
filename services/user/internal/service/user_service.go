package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/user/internal/repository"
)

type UserRepository interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*repository.UserProfile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, fullName string) error
	GetPreferences(ctx context.Context, userID uuid.UUID) (*repository.UserPreferences, error)
	UpsertPreferences(ctx context.Context, prefs *repository.UserPreferences) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

type ProfileResponse struct {
	UserID    string `json:"user_id"`
	FullName  string `json:"full_name"`
	Phone     string `json:"phone,omitempty"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type PreferencesResponse struct {
	RoundToNearest     float64 `json:"round_to_nearest"`
	MaxDailyRoundUp    float64 `json:"max_daily_roundup"`
	Multiplier         int     `json:"multiplier"`
	AutoInvest         bool    `json:"auto_invest"`
	InvestmentStrategy string  `json:"investment_strategy"`
	Language           string  `json:"language"`
	Timezone           string  `json:"timezone"`
}

type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
}

type UpdatePreferencesRequest struct {
	RoundToNearest     *float64 `json:"round_to_nearest,omitempty"`
	MaxDailyRoundUp    *float64 `json:"max_daily_roundup,omitempty"`
	Multiplier         *int     `json:"multiplier,omitempty"`
	AutoInvest         *bool    `json:"auto_invest,omitempty"`
	InvestmentStrategy *string  `json:"investment_strategy,omitempty"`
	Language           *string  `json:"language,omitempty"`
	Timezone           *string  `json:"timezone,omitempty"`
}

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*ProfileResponse, error) {
	p, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	phone := ""
	if p.Phone != nil {
		phone = *p.Phone
	}
	return &ProfileResponse{
		UserID:    p.UserID.String(),
		FullName:  p.FullName,
		Phone:     phone,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) error {
	if req.FullName == "" {
		return fmt.Errorf("full_name is required")
	}
	return s.repo.UpdateProfile(ctx, userID, req.FullName)
}

func (s *UserService) GetPreferences(ctx context.Context, userID uuid.UUID) (*PreferencesResponse, error) {
	p, err := s.repo.GetPreferences(ctx, userID)
	if err != nil {
		return &PreferencesResponse{
			RoundToNearest:     1.00,
			MaxDailyRoundUp:    5.00,
			Multiplier:         1,
			AutoInvest:         true,
			InvestmentStrategy: "moderate",
			Language:           "en",
			Timezone:           "UTC",
		}, nil
	}
	return &PreferencesResponse{
		RoundToNearest:     p.RoundToNearest,
		MaxDailyRoundUp:    p.MaxDailyRoundUp,
		Multiplier:         p.Multiplier,
		AutoInvest:         p.AutoInvest,
		InvestmentStrategy: p.InvestmentStrategy,
		Language:           p.Language,
		Timezone:           p.Timezone,
	}, nil
}

func (s *UserService) UpdatePreferences(ctx context.Context, userID uuid.UUID, req UpdatePreferencesRequest) error {
	prefs := &repository.UserPreferences{
		UserID:             userID,
		RoundToNearest:     1.00,
		MaxDailyRoundUp:    5.00,
		Multiplier:         1,
		AutoInvest:         true,
		InvestmentStrategy: "moderate",
		Language:           "en",
		Timezone:           "UTC",
	}

	existing, err := s.repo.GetPreferences(ctx, userID)
	if err == nil {
		prefs.RoundToNearest = existing.RoundToNearest
		prefs.MaxDailyRoundUp = existing.MaxDailyRoundUp
		prefs.Multiplier = existing.Multiplier
		prefs.AutoInvest = existing.AutoInvest
		prefs.InvestmentStrategy = existing.InvestmentStrategy
		prefs.Language = existing.Language
		prefs.Timezone = existing.Timezone
	}

	if req.RoundToNearest != nil {
		prefs.RoundToNearest = *req.RoundToNearest
	}
	if req.MaxDailyRoundUp != nil {
		prefs.MaxDailyRoundUp = *req.MaxDailyRoundUp
	}
	if req.Multiplier != nil {
		prefs.Multiplier = *req.Multiplier
	}
	if req.AutoInvest != nil {
		prefs.AutoInvest = *req.AutoInvest
	}
	if req.InvestmentStrategy != nil {
		prefs.InvestmentStrategy = *req.InvestmentStrategy
	}
	if req.Language != nil {
		prefs.Language = *req.Language
	}
	if req.Timezone != nil {
		prefs.Timezone = *req.Timezone
	}

	return s.repo.UpsertPreferences(ctx, prefs)
}
