// Copyright (c) 2026 RoundPenny. All rights reserved.

package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/analytics/internal/repository"
)

type mockAnalyticsRepo struct {
	createFn       func(ctx context.Context, e *repository.AnalyticsEvent) error
	listByUserFn   func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repository.AnalyticsEvent, error)
	listByTypeFn   func(ctx context.Context, eventType string, limit, offset int) ([]*repository.AnalyticsEvent, error)
	getDailyStatsFn func(ctx context.Context, startDate, endDate time.Time) ([]*repository.DailyStats, error)
}

func (m *mockAnalyticsRepo) Create(ctx context.Context, e *repository.AnalyticsEvent) error {
	return m.createFn(ctx, e)
}
func (m *mockAnalyticsRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repository.AnalyticsEvent, error) {
	return m.listByUserFn(ctx, userID, limit, offset)
}
func (m *mockAnalyticsRepo) ListByType(ctx context.Context, eventType string, limit, offset int) ([]*repository.AnalyticsEvent, error) {
	return m.listByTypeFn(ctx, eventType, limit, offset)
}
func (m *mockAnalyticsRepo) GetDailyStats(ctx context.Context, startDate, endDate time.Time) ([]*repository.DailyStats, error) {
	return m.getDailyStatsFn(ctx, startDate, endDate)
}

func TestTrackEvent_Success(t *testing.T) {
	eventID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	mock := &mockAnalyticsRepo{
		createFn: func(ctx context.Context, e *repository.AnalyticsEvent) error {
			e.ID = eventID
			e.CreatedAt = now
			return nil
		},
	}

	svc := NewAnalyticsService(mock)
	resp, err := svc.TrackEvent(context.Background(), TrackEventRequest{
		EventType: "page_view",
		UserID:    userID,
		Data:      map[string]any{"page": "/home"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.EventType != "page_view" {
		t.Errorf("expected page_view, got %s", resp.EventType)
	}
}

func TestTrackEvent_InvalidType(t *testing.T) {
	svc := NewAnalyticsService(&mockAnalyticsRepo{})
	_, err := svc.TrackEvent(context.Background(), TrackEventRequest{
		UserID: uuid.New(),
	})
	if err != ErrInvalidType {
		t.Errorf("expected ErrInvalidType, got %v", err)
	}
}

func TestGetUserEvents_DefaultPagination(t *testing.T) {
	userID := uuid.New()

	mock := &mockAnalyticsRepo{
		listByUserFn: func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*repository.AnalyticsEvent, error) {
			if limit != 20 || offset != 0 {
				t.Errorf("expected limit=20, offset=0, got limit=%d offset=%d", limit, offset)
			}
			return []*repository.AnalyticsEvent{}, nil
		},
	}

	svc := NewAnalyticsService(mock)
	resp, err := svc.GetUserEvents(context.Background(), userID, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list")
	}
}

func TestGetDailyStats_Success(t *testing.T) {
	mock := &mockAnalyticsRepo{
		getDailyStatsFn: func(ctx context.Context, startDate, endDate time.Time) ([]*repository.DailyStats, error) {
			return []*repository.DailyStats{
				{Date: "2024-01-01", Count: 10},
				{Date: "2024-01-02", Count: 5},
			}, nil
		},
	}

	svc := NewAnalyticsService(mock)
	start, _ := time.Parse("2006-01-02", "2024-01-01")
	end, _ := time.Parse("2006-01-02", "2024-01-03")
	resp, err := svc.GetDailyStats(context.Background(), start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 stats, got %d", len(resp))
	}
	if resp[0].Count != 10 {
		t.Errorf("expected 10, got %d", resp[0].Count)
	}
}
