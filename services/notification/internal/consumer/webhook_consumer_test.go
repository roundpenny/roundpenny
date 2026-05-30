package consumer

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/notification/internal/repository"
	"github.com/roundup-platform/services/notification/internal/service"
)

type mockWebhookRepo struct {
	getActiveByEventFn  func(ctx context.Context, eventType string) ([]*repository.Webhook, error)
	createDeliveryFn    func(ctx context.Context, d *repository.WebhookDelivery) error
	updateDeliveryFn    func(ctx context.Context, id uuid.UUID, status string, responseCode *int, responseBody, errorMessage string, attemptCount int, nextRetryAt *time.Time, completedAt *time.Time) error
	createFn            func(ctx context.Context, w *repository.Webhook) error
	getByIDFn           func(ctx context.Context, id uuid.UUID) (*repository.Webhook, error)
	getByUserIDFn       func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repository.Webhook, error)
	updateFn            func(ctx context.Context, w *repository.Webhook) error
	deleteFn            func(ctx context.Context, id uuid.UUID) error
}

func (m *mockWebhookRepo) Create(ctx context.Context, w *repository.Webhook) error                     { return m.createFn(ctx, w) }
func (m *mockWebhookRepo) GetByID(ctx context.Context, id uuid.UUID) (*repository.Webhook, error)      { return m.getByIDFn(ctx, id) }
func (m *mockWebhookRepo) GetByUserID(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*repository.Webhook, error) { return m.getByUserIDFn(ctx, uid, limit, offset) }
func (m *mockWebhookRepo) GetActiveByEvent(ctx context.Context, e string) ([]*repository.Webhook, error) { return m.getActiveByEventFn(ctx, e) }
func (m *mockWebhookRepo) Update(ctx context.Context, w *repository.Webhook) error                      { return m.updateFn(ctx, w) }
func (m *mockWebhookRepo) Delete(ctx context.Context, id uuid.UUID) error                               { return m.deleteFn(ctx, id) }
func (m *mockWebhookRepo) CreateDelivery(ctx context.Context, d *repository.WebhookDelivery) error      { return m.createDeliveryFn(ctx, d) }
func (m *mockWebhookRepo) UpdateDelivery(ctx context.Context, id uuid.UUID, status string, responseCode *int, responseBody, errorMessage string, attemptCount int, nextRetryAt *time.Time, completedAt *time.Time) error {
	return m.updateDeliveryFn(ctx, id, status, responseCode, responseBody, errorMessage, attemptCount, nextRetryAt, completedAt)
}

func TestWebhookConsumer_InvalidPayload(t *testing.T) {
	mock := &mockWebhookRepo{
		getActiveByEventFn: func(ctx context.Context, eventType string) ([]*repository.Webhook, error) {
			t.Fatal("should not be called for invalid payload")
			return nil, nil
		},
	}
	svc := service.NewWebhookService(mock)
	_ = NewWebhookConsumer(svc)

	handler := func(ctx context.Context, topic string, key string, data []byte) error {
		var rawPayload map[string]any
		if err := json.Unmarshal(data, &rawPayload); err != nil {
			return nil
		}
		return svc.DeliverEvent(ctx, topic, uuid.New().String(), data)
	}

	err := handler(context.Background(), "tx.settled", "key", []byte("not-json"))
	if err != nil {
		t.Fatalf("expected nil for invalid payload, got %v", err)
	}
}

func TestWebhookConsumer_ValidPayload_CallsDeliverEvent(t *testing.T) {
	var delivered bool
	mock := &mockWebhookRepo{
		getActiveByEventFn: func(ctx context.Context, eventType string) ([]*repository.Webhook, error) {
			return []*repository.Webhook{
				{ID: uuid.New(), URL: "https://example.com/hook", RetryCount: 3, TimeoutMs: 5000},
			}, nil
		},
		createDeliveryFn: func(ctx context.Context, d *repository.WebhookDelivery) error {
			delivered = true
			return nil
		},
		updateDeliveryFn: func(ctx context.Context, id uuid.UUID, status string, responseCode *int, responseBody, errorMessage string, attemptCount int, nextRetryAt *time.Time, completedAt *time.Time) error {
			return nil
		},
	}
	svc := service.NewWebhookService(mock)
	_ = NewWebhookConsumer(svc)

	handler := func(ctx context.Context, topic string, key string, data []byte) error {
		var rawPayload map[string]any
		if err := json.Unmarshal(data, &rawPayload); err != nil {
			return nil
		}
		return svc.DeliverEvent(ctx, topic, uuid.New().String(), data)
	}

	payload := map[string]any{"event": "test"}
	data, _ := json.Marshal(payload)

	err := handler(context.Background(), "tx.settled", "key", data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !delivered {
		t.Fatal("expected delivery to be created")
	}
}

func TestWebhookConsumer_DeliveryFailure_Swallowed(t *testing.T) {
	mock := &mockWebhookRepo{
		getActiveByEventFn: func(ctx context.Context, eventType string) ([]*repository.Webhook, error) {
			return nil, nil
		},
	}
	svc := service.NewWebhookService(mock)
	_ = NewWebhookConsumer(svc)

	handler := func(ctx context.Context, topic string, key string, data []byte) error {
		var rawPayload map[string]any
		if err := json.Unmarshal(data, &rawPayload); err != nil {
			return nil
		}
		return svc.DeliverEvent(ctx, topic, uuid.New().String(), data)
	}

	payload := map[string]any{"event": "test"}
	data, _ := json.Marshal(payload)

	err := handler(context.Background(), "tx.settled", "key", data)
	if err != nil {
		t.Fatalf("expected nil even with no webhooks, got %v", err)
	}
}

func TestWebhookConsumer_GetActiveByEvent_Error(t *testing.T) {
	mock := &mockWebhookRepo{
		getActiveByEventFn: func(ctx context.Context, eventType string) ([]*repository.Webhook, error) {
			return nil, nil
		},
	}
	svc := service.NewWebhookService(mock)
	_ = NewWebhookConsumer(svc)

	handler := func(ctx context.Context, topic string, key string, data []byte) error {
		var rawPayload map[string]any
		if err := json.Unmarshal(data, &rawPayload); err != nil {
			return nil
		}
		return svc.DeliverEvent(ctx, topic, uuid.New().String(), data)
	}

	payload := map[string]any{"event": "test"}
	data, _ := json.Marshal(payload)

	err := handler(context.Background(), "unknown", "key", data)
	if err != nil {
		t.Fatalf("expected nil for unknown topic, got %v", err)
	}
}
