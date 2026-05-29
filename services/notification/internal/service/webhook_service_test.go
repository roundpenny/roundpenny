package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/notification/internal/repository"
)

type mockWebhookRepo struct {
	createFn        func(ctx context.Context, w *repository.Webhook) error
	getByIDFn       func(ctx context.Context, id uuid.UUID) (*repository.Webhook, error)
	getByUserIDFn   func(ctx context.Context, userID uuid.UUID) ([]*repository.Webhook, error)
	getActiveByEventFn func(ctx context.Context, eventType string) ([]*repository.Webhook, error)
	updateFn        func(ctx context.Context, w *repository.Webhook) error
	deleteFn        func(ctx context.Context, id uuid.UUID) error
	createDeliveryFn func(ctx context.Context, d *repository.WebhookDelivery) error
	updateDeliveryFn func(ctx context.Context, id uuid.UUID, status string, responseCode *int, responseBody, errorMessage string, attemptCount int, nextRetryAt *time.Time, completedAt *time.Time) error
}

func (m *mockWebhookRepo) Create(ctx context.Context, w *repository.Webhook) error { return m.createFn(ctx, w) }
func (m *mockWebhookRepo) GetByID(ctx context.Context, id uuid.UUID) (*repository.Webhook, error) { return m.getByIDFn(ctx, id) }
func (m *mockWebhookRepo) GetByUserID(ctx context.Context, uid uuid.UUID) ([]*repository.Webhook, error) { return m.getByUserIDFn(ctx, uid) }
func (m *mockWebhookRepo) GetActiveByEvent(ctx context.Context, e string) ([]*repository.Webhook, error) { return m.getActiveByEventFn(ctx, e) }
func (m *mockWebhookRepo) Update(ctx context.Context, w *repository.Webhook) error { return m.updateFn(ctx, w) }
func (m *mockWebhookRepo) Delete(ctx context.Context, id uuid.UUID) error { return m.deleteFn(ctx, id) }
func (m *mockWebhookRepo) CreateDelivery(ctx context.Context, d *repository.WebhookDelivery) error { return m.createDeliveryFn(ctx, d) }
func (m *mockWebhookRepo) UpdateDelivery(ctx context.Context, id uuid.UUID, status string, responseCode *int, responseBody, errorMessage string, attemptCount int, nextRetryAt *time.Time, completedAt *time.Time) error {
	return m.updateDeliveryFn(ctx, id, status, responseCode, responseBody, errorMessage, attemptCount, nextRetryAt, completedAt)
}

func TestCreateWebhook_Success(t *testing.T) {
	userID := uuid.New()
	hookID := uuid.New()
	now := time.Now()

	mock := &mockWebhookRepo{
		createFn: func(ctx context.Context, w *repository.Webhook) error {
			w.ID = hookID
			w.CreatedAt = now
			w.UpdatedAt = now
			return nil
		},
	}

	svc := NewWebhookService(mock)
	resp, err := svc.CreateWebhook(context.Background(), CreateWebhookRequest{
		UserID:      userID,
		URL:         "https://example.com/webhook",
		Events:      []string{"tx.settled", "roundup.calculated"},
		Description: "test hook",
		RetryCount:  3,
		TimeoutMs:   5000,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.URL != "https://example.com/webhook" {
		t.Errorf("bad url: %s", resp.URL)
	}
	if len(resp.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(resp.Events))
	}
}

func TestCreateWebhook_EmptyURL(t *testing.T) {
	svc := NewWebhookService(&mockWebhookRepo{})
	_, err := svc.CreateWebhook(context.Background(), CreateWebhookRequest{
		UserID: uuid.New(),
		URL:    "",
		Events: []string{"tx.settled"},
	})
	if err == nil {
		t.Fatal("expected error for empty url")
	}
}

func TestCreateWebhook_DefaultRetryAndTimeout(t *testing.T) {
	now := time.Now()
	mock := &mockWebhookRepo{
		createFn: func(ctx context.Context, w *repository.Webhook) error {
			w.ID = uuid.New()
			w.CreatedAt = now
			w.UpdatedAt = now
			return nil
		},
	}

	svc := NewWebhookService(mock)
	resp, err := svc.CreateWebhook(context.Background(), CreateWebhookRequest{
		UserID: uuid.New(),
		URL:    "https://example.com/hook",
		Events: []string{"tx.settled"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.RetryCount != 3 {
		t.Errorf("expected 3 retries, got %d", resp.RetryCount)
	}
	if resp.TimeoutMs != 5000 {
		t.Errorf("expected 5000ms timeout, got %d", resp.TimeoutMs)
	}
}

func TestGetWebhook_NotFound(t *testing.T) {
	mock := &mockWebhookRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Webhook, error) {
			return nil, nil
		},
	}

	svc := NewWebhookService(mock)
	_, err := svc.GetWebhook(context.Background(), uuid.New())
	if err != ErrWebhookNotFound {
		t.Errorf("expected ErrWebhookNotFound, got %v", err)
	}
}

func TestUpdateWebhook_Success(t *testing.T) {
	hookID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	currentURL := "https://old-url.com/hook"

	mock := &mockWebhookRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Webhook, error) {
			return &repository.Webhook{
				ID:     hookID,
				UserID: userID,
				URL:    currentURL,
				Events: []string{"tx.settled"},
			}, nil
		},
		updateFn: func(ctx context.Context, w *repository.Webhook) error {
			currentURL = w.URL
			w.UpdatedAt = now
			return nil
		},
	}

	svc := NewWebhookService(mock)
	active := true
	resp, err := svc.UpdateWebhook(context.Background(), hookID, UpdateWebhookRequest{
		URL:      "https://new-url.com/hook",
		IsActive: &active,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.URL != "https://new-url.com/hook" {
		t.Errorf("expected new url, got %s", resp.URL)
	}
}

func TestDeleteWebhook_Success(t *testing.T) {
	deleted := false
	mock := &mockWebhookRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.Webhook, error) {
			return &repository.Webhook{ID: id}, nil
		},
		deleteFn: func(ctx context.Context, id uuid.UUID) error {
			deleted = true
			return nil
		},
	}

	svc := NewWebhookService(mock)
	if err := svc.DeleteWebhook(context.Background(), uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatal("delete was not called")
	}
}

func TestDeliverEvent_FiltersByEventType(t *testing.T) {
	mock := &mockWebhookRepo{
		getActiveByEventFn: func(ctx context.Context, eventType string) ([]*repository.Webhook, error) {
			if eventType != "tx.settled" {
				t.Errorf("expected tx.settled, got %s", eventType)
			}
			return []*repository.Webhook{}, nil
		},
	}

	svc := NewWebhookService(mock)
	if err := svc.DeliverEvent(context.Background(), "tx.settled", "evt_1", []byte(`{"hello":"world"}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
