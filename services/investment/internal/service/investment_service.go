package service

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/roundup-platform/pkg/event"
	"github.com/roundup-platform/pkg/kafka"
	"github.com/roundup-platform/services/investment/internal/repository"
)

type InvestmentRepository interface {
	GetOrCreatePortfolio(ctx context.Context, userID uuid.UUID, strategy string) (*repository.Portfolio, error)
	CreateInvestment(ctx context.Context, inv *repository.Investment) error
	AddToPortfolioBalance(ctx context.Context, portfolioID uuid.UUID, amount float64) error
}

type InvestmentProducer interface {
	Publish(ctx context.Context, msg kafka.Message) error
}

type InvestmentService struct {
	repo     InvestmentRepository
	producer InvestmentProducer
}

func NewInvestmentService(repo InvestmentRepository, producer InvestmentProducer) *InvestmentService {
	return &InvestmentService{repo: repo, producer: producer}
}

func (s *InvestmentService) InvestRoundUp(ctx context.Context, userID uuid.UUID, amount float64) error {
	portfolio, err := s.repo.GetOrCreatePortfolio(ctx, userID, "moderate")
	if err != nil {
		return fmt.Errorf("get portfolio: %w", err)
	}

	netAmount := amount * 0.90

	inv := &repository.Investment{
		PortfolioID: portfolio.ID,
		UserID:      userID,
		Amount:      netAmount,
		Source:      "roundup",
		Status:      "invested",
	}

	if err := s.repo.CreateInvestment(ctx, inv); err != nil {
		return fmt.Errorf("create investment: %w", err)
	}

	if err := s.repo.AddToPortfolioBalance(ctx, portfolio.ID, netAmount); err != nil {
		return fmt.Errorf("update balance: %w", err)
	}

	evt := event.InvestmentCreated{
		UserID:    userID.String(),
		Portfolio: portfolio.ID.String(),
		Amount:    netAmount,
	}

	if err := s.producer.Publish(ctx, kafka.Message{
		Key:     userID.String(),
		Topic:   event.TopicInvestmentCreated,
		Payload: evt,
	}); err != nil {
		log.Printf("investment publish warning: %v", err)
	}

	log.Printf("invested: user=%s amount=%.2f portfolio=%s",
		userID, netAmount, portfolio.ID)

	return nil
}
