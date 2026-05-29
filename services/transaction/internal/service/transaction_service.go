package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/transaction/internal/repository"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *repository.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*repository.Transaction, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]repository.Transaction, error)
	Settle(ctx context.Context, id uuid.UUID) error
}

type TransactionProducer interface {
	Publish(ctx context.Context, msg kafka.Message) error
}

type TransactionService struct {
	repo     TransactionRepository
	producer TransactionProducer
}

func NewTransactionService(repo TransactionRepository, producer TransactionProducer) *TransactionService {
	return &TransactionService{repo: repo, producer: producer}
}

type CreateTransactionRequest struct {
	UserID           string  `json:"user_id"`
	MerchantID       string  `json:"merchant_id,omitempty"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	Type             string  `json:"type"`
	ExternalTxID     string  `json:"external_tx_id"`
	ExternalProvider string  `json:"external_provider"`
	Description      string  `json:"description"`
}

func (s *TransactionService) Create(ctx context.Context, req CreateTransactionRequest) (*repository.Transaction, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	tx := &repository.Transaction{
		UserID:           userID,
		Amount:           req.Amount,
		Currency:         req.Currency,
		Type:             req.Type,
		ExternalTxID:     req.ExternalTxID,
		ExternalProvider: req.ExternalProvider,
		Description:      req.Description,
	}

	if req.MerchantID != "" {
		mid, err := uuid.Parse(req.MerchantID)
		if err != nil {
			return nil, fmt.Errorf("invalid merchant_id: %w", err)
		}
		tx.MerchantID = &mid
	}

	if err := s.repo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	if err := s.SettleAndPublish(ctx, tx.ID); err != nil {
		return nil, fmt.Errorf("settle: %w", err)
	}

	return tx, nil
}

func (s *TransactionService) SettleAndPublish(ctx context.Context, txID uuid.UUID) error {
	if err := s.repo.Settle(ctx, txID); err != nil {
		return fmt.Errorf("settle: %w", err)
	}

	tx, err := s.repo.GetByID(ctx, txID)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}

	evt := event.TransactionSettled{
		TransactionID: tx.ID.String(),
		UserID:        tx.UserID.String(),
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		ExternalTxID:  tx.ExternalTxID,
	}

	if tx.MerchantID != nil {
		evt.MerchantID = tx.MerchantID.String()
	}

	if err := s.producer.Publish(ctx, kafka.Message{
		Key:     tx.UserID.String(),
		Topic:   event.TopicTransactionSettled,
		Payload: evt,
	}); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	return nil
}

func (s *TransactionService) GetByID(ctx context.Context, id uuid.UUID) (*repository.Transaction, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TransactionService) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]repository.Transaction, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListByUser(ctx, userID, pageSize, offset)
}

func (s *TransactionService) ProcessWebhook(ctx context.Context, provider string, payload []byte) (*repository.Transaction, error) {
	now := time.Now()
	tx := &repository.Transaction{
		Amount:           0,
		Currency:         "USD",
		Type:             "purchase",
		ExternalProvider: provider,
		ExternalTxID:     fmt.Sprintf("%s-%d", provider, now.UnixNano()),
		Status:           "pending",
	}

	return tx, nil
}
