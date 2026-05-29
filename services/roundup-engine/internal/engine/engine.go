package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/roundup-engine/internal/repository"
)

type RoundUpRepository interface {
	GetUserPreferences(ctx context.Context, userID uuid.UUID) (*repository.UserPreference, error)
	GetDailyRoundUpTotal(ctx context.Context, userID uuid.UUID) (float64, error)
	CreateRoundUp(ctx context.Context, ru *repository.RoundUpRecord) error
	UpdateRoundUpStatus(ctx context.Context, id uuid.UUID, status string) error
}

type RoundUpProducer interface {
	Publish(ctx context.Context, msg kafka.Message) error
}

type RoundUpEngine struct {
	repo     RoundUpRepository
	producer RoundUpProducer
}

func NewRoundUpEngine(repo RoundUpRepository, producer RoundUpProducer) *RoundUpEngine {
	return &RoundUpEngine{repo: repo, producer: producer}
}

func (e *RoundUpEngine) HandleTransaction(ctx context.Context, topic string, key string, data []byte) error {
	var tx event.TransactionSettled
	if err := json.Unmarshal(data, &tx); err != nil {
		return fmt.Errorf("unmarshal transaction: %w", err)
	}

	userID, err := uuid.Parse(tx.UserID)
	if err != nil {
		return fmt.Errorf("parse user_id: %w", err)
	}

	prefs, err := e.repo.GetUserPreferences(ctx, userID)
	if err != nil {
		return fmt.Errorf("get preferences: %w", err)
	}

	dailyTotal, err := e.repo.GetDailyRoundUpTotal(ctx, userID)
	if err != nil {
		return fmt.Errorf("get daily total: %w", err)
	}

	roundUp := Calculate(tx.Amount, prefs.RoundToNearest, prefs.Multiplier)
	if roundUp <= 0 {
		return nil
	}

	if dailyTotal+roundUp > prefs.MaxDailyRoundup {
		roundUp = math.Max(0, prefs.MaxDailyRoundup-dailyTotal)
		if roundUp <= 0 {
			return nil
		}
	}

	txID, err := uuid.Parse(tx.TransactionID)
	if err != nil {
		return fmt.Errorf("parse transaction_id: %w", err)
	}

	roundedAmount := tx.Amount + roundUp

	record := &repository.RoundUpRecord{
		TransactionID:   txID,
		UserID:          userID,
		OriginalAmount:  tx.Amount,
		RoundedAmount:   roundedAmount,
		RoundUpAmount:   roundUp,
		Currency:        tx.Currency,
		Status:          "pending",
	}

	if err := e.repo.CreateRoundUp(ctx, record); err != nil {
		return fmt.Errorf("create roundup: %w", err)
	}

	evt := event.RoundUpCalculated{
		TransactionID:  tx.TransactionID,
		UserID:         tx.UserID,
		OriginalAmount: tx.Amount,
		RoundedAmount:  roundedAmount,
		RoundUpAmount:  roundUp,
		Currency:       tx.Currency,
	}

	if err := e.producer.Publish(ctx, kafka.Message{
		Key:     tx.UserID,
		Topic:   event.TopicRoundUpCalculated,
		Payload: evt,
	}); err != nil {
		e.repo.UpdateRoundUpStatus(ctx, record.ID, "failed")
		return fmt.Errorf("publish roundup: %w", err)
	}

	log.Printf("roundup: user=%s tx=%s amount=%.2f -> %.2f (roundup=%.2f)",
		tx.UserID, tx.TransactionID, tx.Amount, roundedAmount, roundUp)

	return nil
}

func Calculate(amount, roundToNearest float64, multiplier int) float64 {
	if roundToNearest <= 0 {
		roundToNearest = 1.00
	}
	if multiplier <= 0 {
		multiplier = 1
	}

	remainder := math.Mod(amount, roundToNearest)
	if remainder == 0 {
		return 0
	}

	roundUp := roundToNearest - remainder
	return roundUp * float64(multiplier)
}
