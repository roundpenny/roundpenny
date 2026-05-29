package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/analytics/internal/repository"
)

var (
	ErrEventNotFound = errors.New("analytics event not found")
	ErrInvalidType   = errors.New("event type is required")
)

type AnalyticsRepository interface {
	Create(ctx context.Context, e *repository.AnalyticsEvent) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repository.AnalyticsEvent, error)
	ListByType(ctx context.Context, eventType string, limit, offset int) ([]*repository.AnalyticsEvent, error)
	GetDailyStats(ctx context.Context, startDate, endDate time.Time) ([]*repository.DailyStats, error)
}

type TrackEventRequest struct {
	EventType string         `json:"event_type"`
	UserID    uuid.UUID      `json:"user_id"`
	Data      map[string]any `json:"data,omitempty"`
}

type AnalyticsEventResponse struct {
	ID        uuid.UUID              `json:"id"`
	EventType string                 `json:"event_type"`
	UserID    uuid.UUID              `json:"user_id"`
	Data      map[string]any         `json:"data,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type DailyStatsResponse struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type AnalyticsService struct {
	repo AnalyticsRepository
}

func NewAnalyticsService(repo AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) TrackEvent(ctx context.Context, req TrackEventRequest) (*AnalyticsEventResponse, error) {
	if req.EventType == "" {
		return nil, ErrInvalidType
	}

	e := &repository.AnalyticsEvent{
		EventType: req.EventType,
		UserID:    req.UserID,
		Data:      req.Data,
	}

	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}

	return toResponse(e), nil
}

func (s *AnalyticsService) GetUserEvents(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*AnalyticsEventResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	events, err := s.repo.ListByUser(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*AnalyticsEventResponse, len(events))
	for i, e := range events {
		responses[i] = toResponse(e)
	}
	return responses, nil
}

func (s *AnalyticsService) GetDailyStats(ctx context.Context, startDate, endDate time.Time) ([]*DailyStatsResponse, error) {
	stats, err := s.repo.GetDailyStats(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	responses := make([]*DailyStatsResponse, len(stats))
	for i, s := range stats {
		responses[i] = &DailyStatsResponse{
			Date:  s.Date,
			Count: s.Count,
		}
	}
	return responses, nil
}

func toResponse(e *repository.AnalyticsEvent) *AnalyticsEventResponse {
	return &AnalyticsEventResponse{
		ID:        e.ID,
		EventType: e.EventType,
		UserID:    e.UserID,
		Data:      e.Data,
		CreatedAt: e.CreatedAt,
	}
}
