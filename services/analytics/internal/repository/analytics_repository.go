package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/db"
)

type AnalyticsEvent struct {
	ID        uuid.UUID              `json:"id"`
	EventType string                 `json:"event_type"`
	UserID    uuid.UUID              `json:"user_id"`
	Data      map[string]any         `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

type DailyStats struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type AnalyticsRepository struct {
	pool *db.Pool
}

func NewAnalyticsRepository(pool *db.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

func (r *AnalyticsRepository) Create(ctx context.Context, e *AnalyticsEvent) error {
	query := `INSERT INTO analytics_events (event_type, user_id, data)
	          VALUES ($1, $2, $3)
	          RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, e.EventType, e.UserID, e.Data).Scan(&e.ID, &e.CreatedAt)
}

func (r *AnalyticsRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AnalyticsEvent, error) {
	query := `SELECT id, event_type, user_id, data, created_at
	          FROM analytics_events WHERE user_id = $1
	          ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*AnalyticsEvent
	for rows.Next() {
		e := &AnalyticsEvent{}
		var dataJSON []byte
		if err := rows.Scan(&e.ID, &e.EventType, &e.UserID, &dataJSON, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *AnalyticsRepository) ListByType(ctx context.Context, eventType string, limit, offset int) ([]*AnalyticsEvent, error) {
	query := `SELECT id, event_type, user_id, data, created_at
	          FROM analytics_events WHERE event_type = $1
	          ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, eventType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*AnalyticsEvent
	for rows.Next() {
		e := &AnalyticsEvent{}
		var dataJSON []byte
		if err := rows.Scan(&e.ID, &e.EventType, &e.UserID, &dataJSON, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *AnalyticsRepository) GetDailyStats(ctx context.Context, startDate, endDate time.Time) ([]*DailyStats, error) {
	query := `SELECT DATE(created_at)::TEXT AS date, COUNT(*) AS count
	          FROM analytics_events
	          WHERE created_at >= $1 AND created_at < $2
	          GROUP BY DATE(created_at)
	          ORDER BY date ASC`
	rows, err := r.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*DailyStats
	for rows.Next() {
		s := &DailyStats{}
		if err := rows.Scan(&s.Date, &s.Count); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}
