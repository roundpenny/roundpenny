package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/notification/internal/repository"
)

var (
	ErrWebhookNotFound = errors.New("webhook not found")
)

type WebhookRepository interface {
	Create(ctx context.Context, w *repository.Webhook) error
	GetByID(ctx context.Context, id uuid.UUID) (*repository.Webhook, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*repository.Webhook, error)
	GetActiveByEvent(ctx context.Context, eventType string) ([]*repository.Webhook, error)
	Update(ctx context.Context, w *repository.Webhook) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateDelivery(ctx context.Context, d *repository.WebhookDelivery) error
	UpdateDelivery(ctx context.Context, id uuid.UUID, status string, responseCode *int, responseBody, errorMessage string, attemptCount int, nextRetryAt *time.Time, completedAt *time.Time) error
}

type CreateWebhookRequest struct {
	UserID      uuid.UUID
	URL         string
	Secret      string
	Events      []string
	Description string
	RetryCount  int
	TimeoutMs   int
}

type UpdateWebhookRequest struct {
	URL         string
	Secret      string
	Events      []string
	IsActive    *bool
	Description string
	RetryCount  int
	TimeoutMs   int
}

type WebhookResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	URL         string    `json:"url"`
	Events      []string  `json:"events"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description,omitempty"`
	RetryCount  int       `json:"retry_count"`
	TimeoutMs   int       `json:"timeout_ms"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WebhookService struct {
	repo    WebhookRepository
	client  *http.Client
}

func NewWebhookService(repo WebhookRepository) *WebhookService {
	return &WebhookService{
		repo: repo,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,
			},
		},
	}
}

func (s *WebhookService) CreateWebhook(ctx context.Context, req CreateWebhookRequest) (*WebhookResponse, error) {
	if req.URL == "" {
		return nil, errors.New("url is required")
	}
	if len(req.Events) == 0 {
		return nil, errors.New("at least one event is required")
	}
	if req.RetryCount <= 0 {
		req.RetryCount = 3
	}
	if req.TimeoutMs <= 0 {
		req.TimeoutMs = 5000
	}

	w := &repository.Webhook{
		UserID:      req.UserID,
		URL:         req.URL,
		Secret:      req.Secret,
		Events:      req.Events,
		IsActive:    true,
		Description: req.Description,
		RetryCount:  req.RetryCount,
		TimeoutMs:   req.TimeoutMs,
	}

	if err := s.repo.Create(ctx, w); err != nil {
		return nil, err
	}

	return webhookToResponse(w), nil
}

func (s *WebhookService) GetWebhook(ctx context.Context, id uuid.UUID) (*WebhookResponse, error) {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, ErrWebhookNotFound
	}
	return webhookToResponse(w), nil
}

func (s *WebhookService) ListUserWebhooks(ctx context.Context, userID uuid.UUID) ([]*WebhookResponse, error) {
	webhooks, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]*WebhookResponse, len(webhooks))
	for i, w := range webhooks {
		responses[i] = webhookToResponse(w)
	}
	return responses, nil
}

func (s *WebhookService) UpdateWebhook(ctx context.Context, id uuid.UUID, req UpdateWebhookRequest) (*WebhookResponse, error) {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, ErrWebhookNotFound
	}

	if req.URL != "" {
		w.URL = req.URL
	}
	if len(req.Events) > 0 {
		w.Events = req.Events
	}
	if req.IsActive != nil {
		w.IsActive = *req.IsActive
	}
	w.Secret = req.Secret
	w.Description = req.Description
	w.RetryCount = req.RetryCount
	w.TimeoutMs = req.TimeoutMs

	if err := s.repo.Update(ctx, w); err != nil {
		return nil, err
	}

	return s.GetWebhook(ctx, id)
}

func (s *WebhookService) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if w == nil {
		return ErrWebhookNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *WebhookService) DeliverEvent(ctx context.Context, eventType string, eventID string, payload []byte) error {
	webhooks, err := s.repo.GetActiveByEvent(ctx, eventType)
	if err != nil {
		return fmt.Errorf("get webhooks for event: %w", err)
	}

	for _, w := range webhooks {
		now := time.Now()
		d := &repository.WebhookDelivery{
			WebhookID:   w.ID,
			EventType:   eventType,
			EventID:     eventID,
			Payload:     payload,
			Status:      "pending",
			MaxAttempts: w.RetryCount,
			NextRetryAt: &now,
		}

		if err := s.repo.CreateDelivery(ctx, d); err != nil {
			log.Printf("create delivery record: %v", err)
			continue
		}

		go s.deliverWithRetry(d.ID, w, payload)
	}

	return nil
}

func (s *WebhookService) deliverWithRetry(deliveryID uuid.UUID, w *repository.Webhook, payload []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(w.TimeoutMs)*time.Millisecond)
	defer cancel()

	for attempt := 1; attempt <= w.RetryCount; attempt++ {
		statusCode, respBody, err := s.sendWebhook(ctx, w.URL, w.Secret, payload)

		if err == nil && statusCode >= 200 && statusCode < 300 {
			now := time.Now()
			_ = s.repo.UpdateDelivery(ctx, deliveryID, "delivered", &statusCode, respBody, "", attempt, nil, &now)
			return
		}

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		} else {
			errMsg = fmt.Sprintf("HTTP %d", statusCode)
		}

		if attempt < w.RetryCount {
			backoff := time.Duration(attempt*attempt) * time.Second
			nextRetry := time.Now().Add(backoff)
			_ = s.repo.UpdateDelivery(ctx, deliveryID, "retrying", &statusCode, respBody, errMsg, attempt, &nextRetry, nil)
			time.Sleep(backoff)
		} else {
			_ = s.repo.UpdateDelivery(ctx, deliveryID, "failed", &statusCode, respBody, errMsg, attempt, nil, nil)
		}
	}
}

func (s *WebhookService) sendWebhook(ctx context.Context, url, secret string, payload []byte) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Roundup-Webhook/1.0")
	if secret != "" {
		req.Header.Set("X-Webhook-Signature", secret)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body), nil
}

func webhookToResponse(w *repository.Webhook) *WebhookResponse {
	return &WebhookResponse{
		ID:          w.ID,
		UserID:      w.UserID,
		URL:         w.URL,
		Events:      w.Events,
		IsActive:    w.IsActive,
		Description: w.Description,
		RetryCount:  w.RetryCount,
		TimeoutMs:   w.TimeoutMs,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}
