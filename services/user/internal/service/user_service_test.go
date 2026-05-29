package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/user/internal/repository"
)

type mockUserRepo struct {
	getProfileFn        func(ctx context.Context, userID uuid.UUID) (*repository.UserProfile, error)
	updateProfileFn     func(ctx context.Context, userID uuid.UUID, fullName string) error
	getPreferencesFn    func(ctx context.Context, userID uuid.UUID) (*repository.UserPreferences, error)
	upsertPreferencesFn func(ctx context.Context, prefs *repository.UserPreferences) error
}

func (m *mockUserRepo) GetProfile(ctx context.Context, userID uuid.UUID) (*repository.UserProfile, error) {
	return m.getProfileFn(ctx, userID)
}

func (m *mockUserRepo) UpdateProfile(ctx context.Context, userID uuid.UUID, fullName string) error {
	return m.updateProfileFn(ctx, userID, fullName)
}

func (m *mockUserRepo) GetPreferences(ctx context.Context, userID uuid.UUID) (*repository.UserPreferences, error) {
	return m.getPreferencesFn(ctx, userID)
}

func (m *mockUserRepo) UpsertPreferences(ctx context.Context, prefs *repository.UserPreferences) error {
	return m.upsertPreferencesFn(ctx, prefs)
}

func TestGetProfile_Success(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	svc := NewUserService(&mockUserRepo{
		getProfileFn: func(ctx context.Context, uid uuid.UUID) (*repository.UserProfile, error) {
			return &repository.UserProfile{
				UserID:    uid,
				FullName:  "John Doe",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	})

	resp, err := svc.GetProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UserID != userID.String() {
		t.Errorf("expected user ID %s, got %s", userID.String(), resp.UserID)
	}
	if resp.FullName != "John Doe" {
		t.Errorf("expected John Doe, got %s", resp.FullName)
	}
	if resp.CreatedAt == "" {
		t.Errorf("expected non-empty created_at")
	}
}

func TestGetProfile_WithPhone(t *testing.T) {
	userID := uuid.New()
	phone := "+1234567890"
	now := time.Now()

	svc := NewUserService(&mockUserRepo{
		getProfileFn: func(ctx context.Context, uid uuid.UUID) (*repository.UserProfile, error) {
			return &repository.UserProfile{
				UserID:    uid,
				FullName:  "Jane Doe",
				Phone:     &phone,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	})

	resp, err := svc.GetProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Phone != phone {
		t.Errorf("expected phone %s, got %s", phone, resp.Phone)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	svc := NewUserService(&mockUserRepo{
		getProfileFn: func(ctx context.Context, uid uuid.UUID) (*repository.UserProfile, error) {
			return nil, errors.New("not found")
		},
	})

	_, err := svc.GetProfile(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for not found profile")
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	var capturedName string

	svc := NewUserService(&mockUserRepo{
		updateProfileFn: func(ctx context.Context, uid uuid.UUID, fullName string) error {
			capturedName = fullName
			return nil
		},
	})

	err := svc.UpdateProfile(context.Background(), uuid.New(), UpdateProfileRequest{FullName: "Jane Doe"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedName != "Jane Doe" {
		t.Errorf("expected Jane Doe, got %s", capturedName)
	}
}

func TestUpdateProfile_MissingFullName(t *testing.T) {
	svc := NewUserService(&mockUserRepo{})

	err := svc.UpdateProfile(context.Background(), uuid.New(), UpdateProfileRequest{FullName: ""})
	if err == nil {
		t.Fatal("expected error for missing full_name")
	}
}

func TestGetPreferences_Success(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	svc := NewUserService(&mockUserRepo{
		getPreferencesFn: func(ctx context.Context, uid uuid.UUID) (*repository.UserPreferences, error) {
			return &repository.UserPreferences{
				UserID:             uid,
				RoundToNearest:     1.00,
				MaxDailyRoundUp:    5.00,
				Multiplier:         1,
				AutoInvest:         true,
				InvestmentStrategy: "aggressive",
				Language:           "en",
				Timezone:           "UTC",
				CreatedAt:          now,
				UpdatedAt:          now,
			}, nil
		},
	})

	resp, err := svc.GetPreferences(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.InvestmentStrategy != "aggressive" {
		t.Errorf("expected aggressive, got %s", resp.InvestmentStrategy)
	}
	if resp.Language != "en" {
		t.Errorf("expected en, got %s", resp.Language)
	}
	if resp.RoundToNearest != 1.00 {
		t.Errorf("expected 1.00, got %f", resp.RoundToNearest)
	}
}

func TestGetPreferences_NoExisting(t *testing.T) {
	svc := NewUserService(&mockUserRepo{
		getPreferencesFn: func(ctx context.Context, uid uuid.UUID) (*repository.UserPreferences, error) {
			return nil, errors.New("not found")
		},
	})

	resp, err := svc.GetPreferences(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RoundToNearest != 1.00 {
		t.Errorf("expected default RoundToNearest 1.00, got %f", resp.RoundToNearest)
	}
	if resp.MaxDailyRoundUp != 5.00 {
		t.Errorf("expected default MaxDailyRoundUp 5.00, got %f", resp.MaxDailyRoundUp)
	}
	if resp.Multiplier != 1 {
		t.Errorf("expected default Multiplier 1, got %d", resp.Multiplier)
	}
	if !resp.AutoInvest {
		t.Errorf("expected default AutoInvest true")
	}
	if resp.InvestmentStrategy != "moderate" {
		t.Errorf("expected default InvestmentStrategy moderate, got %s", resp.InvestmentStrategy)
	}
	if resp.Language != "en" {
		t.Errorf("expected default Language en, got %s", resp.Language)
	}
	if resp.Timezone != "UTC" {
		t.Errorf("expected default Timezone UTC, got %s", resp.Timezone)
	}
}

func TestUpdatePreferences_Success(t *testing.T) {
	userID := uuid.New()
	multiplier := 2
	lang := "fr"

	svc := NewUserService(&mockUserRepo{
		getPreferencesFn: func(ctx context.Context, uid uuid.UUID) (*repository.UserPreferences, error) {
			return &repository.UserPreferences{
				UserID:             uid,
				RoundToNearest:     1.00,
				MaxDailyRoundUp:    5.00,
				Multiplier:         1,
				AutoInvest:         true,
				InvestmentStrategy: "moderate",
				Language:           "en",
				Timezone:           "UTC",
			}, nil
		},
		upsertPreferencesFn: func(ctx context.Context, prefs *repository.UserPreferences) error {
			if prefs.Multiplier != 2 {
				t.Errorf("expected multiplier 2, got %d", prefs.Multiplier)
			}
			if prefs.Language != "fr" {
				t.Errorf("expected language fr, got %s", prefs.Language)
			}
			return nil
		},
	})

	err := svc.UpdatePreferences(context.Background(), userID, UpdatePreferencesRequest{
		Multiplier: &multiplier,
		Language:   &lang,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdatePreferences_NoExisting(t *testing.T) {
	userID := uuid.New()

	svc := NewUserService(&mockUserRepo{
		getPreferencesFn: func(ctx context.Context, uid uuid.UUID) (*repository.UserPreferences, error) {
			return nil, errors.New("not found")
		},
		upsertPreferencesFn: func(ctx context.Context, prefs *repository.UserPreferences) error {
			if prefs.RoundToNearest != 1.00 {
				t.Errorf("expected default RoundToNearest 1.00, got %f", prefs.RoundToNearest)
			}
			if prefs.MaxDailyRoundUp != 5.00 {
				t.Errorf("expected default MaxDailyRoundUp 5.00, got %f", prefs.MaxDailyRoundUp)
			}
			if prefs.Multiplier != 1 {
				t.Errorf("expected default Multiplier 1, got %d", prefs.Multiplier)
			}
			if !prefs.AutoInvest {
				t.Errorf("expected default AutoInvest true")
			}
			if prefs.InvestmentStrategy != "moderate" {
				t.Errorf("expected default InvestmentStrategy moderate, got %s", prefs.InvestmentStrategy)
			}
			if prefs.Language != "en" {
				t.Errorf("expected default Language en, got %s", prefs.Language)
			}
			if prefs.Timezone != "UTC" {
				t.Errorf("expected default Timezone UTC, got %s", prefs.Timezone)
			}
			return nil
		},
	})

	err := svc.UpdatePreferences(context.Background(), userID, UpdatePreferencesRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdatePreferences_AllFields(t *testing.T) {
	userID := uuid.New()
	rt := 2.00
	md := 10.00
	m := 3
	ai := false
	is := "conservative"
	lang := "de"
	tz := "Europe/Berlin"

	svc := NewUserService(&mockUserRepo{
		getPreferencesFn: func(ctx context.Context, uid uuid.UUID) (*repository.UserPreferences, error) {
			return nil, errors.New("not found")
		},
		upsertPreferencesFn: func(ctx context.Context, prefs *repository.UserPreferences) error {
			if prefs.RoundToNearest != 2.00 {
				t.Errorf("expected RoundToNearest 2.00, got %f", prefs.RoundToNearest)
			}
			if prefs.MaxDailyRoundUp != 10.00 {
				t.Errorf("expected MaxDailyRoundUp 10.00, got %f", prefs.MaxDailyRoundUp)
			}
			if prefs.Multiplier != 3 {
				t.Errorf("expected Multiplier 3, got %d", prefs.Multiplier)
			}
			if prefs.AutoInvest != false {
				t.Errorf("expected AutoInvest false, got %v", prefs.AutoInvest)
			}
			if prefs.InvestmentStrategy != "conservative" {
				t.Errorf("expected InvestmentStrategy conservative, got %s", prefs.InvestmentStrategy)
			}
			if prefs.Language != "de" {
				t.Errorf("expected Language de, got %s", prefs.Language)
			}
			if prefs.Timezone != "Europe/Berlin" {
				t.Errorf("expected Timezone Europe/Berlin, got %s", prefs.Timezone)
			}
			return nil
		},
	})

	err := svc.UpdatePreferences(context.Background(), userID, UpdatePreferencesRequest{
		RoundToNearest:     &rt,
		MaxDailyRoundUp:    &md,
		Multiplier:         &m,
		AutoInvest:         &ai,
		InvestmentStrategy: &is,
		Language:           &lang,
		Timezone:           &tz,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdatePreferences_PartialUpdate(t *testing.T) {
	userID := uuid.New()
	tz := "America/New_York"

	svc := NewUserService(&mockUserRepo{
		getPreferencesFn: func(ctx context.Context, uid uuid.UUID) (*repository.UserPreferences, error) {
			return &repository.UserPreferences{
				UserID:             uid,
				RoundToNearest:     0.50,
				MaxDailyRoundUp:    10.00,
				Multiplier:         3,
				AutoInvest:         false,
				InvestmentStrategy: "aggressive",
				Language:           "es",
				Timezone:           "UTC",
			}, nil
		},
		upsertPreferencesFn: func(ctx context.Context, prefs *repository.UserPreferences) error {
			if prefs.RoundToNearest != 0.50 {
				t.Errorf("expected existing RoundToNearest 0.50, got %f", prefs.RoundToNearest)
			}
			if prefs.MaxDailyRoundUp != 10.00 {
				t.Errorf("expected existing MaxDailyRoundUp 10.00, got %f", prefs.MaxDailyRoundUp)
			}
			if prefs.Multiplier != 3 {
				t.Errorf("expected existing Multiplier 3, got %d", prefs.Multiplier)
			}
			if prefs.AutoInvest != false {
				t.Errorf("expected existing AutoInvest false, got %v", prefs.AutoInvest)
			}
			if prefs.InvestmentStrategy != "aggressive" {
				t.Errorf("expected existing InvestmentStrategy aggressive, got %s", prefs.InvestmentStrategy)
			}
			if prefs.Language != "es" {
				t.Errorf("expected existing Language es, got %s", prefs.Language)
			}
			if prefs.Timezone != "America/New_York" {
				t.Errorf("expected Timezone America/New_York, got %s", prefs.Timezone)
			}
			return nil
		},
	})

	err := svc.UpdatePreferences(context.Background(), userID, UpdatePreferencesRequest{
		Timezone: &tz,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
