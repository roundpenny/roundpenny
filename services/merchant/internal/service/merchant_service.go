// Copyright (c) 2026 RoundPenny. All rights reserved.

package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/merchant/internal/repository"
)

var (
	ErrMerchantNotFound = errors.New("merchant not found")
	ErrInvalidName      = errors.New("merchant name is required")
	ErrInvalidEmail     = errors.New("valid email is required")
)

type MerchantRepository interface {
	Create(ctx context.Context, m *repository.Merchant) error
	GetByID(ctx context.Context, id uuid.UUID) (*repository.Merchant, error)
	Update(ctx context.Context, m *repository.Merchant) error
	Delete(ctx context.Context, id uuid.UUID) error
	Search(ctx context.Context, query string, limit, offset int) ([]*repository.Merchant, error)
	List(ctx context.Context, limit, offset int) ([]*repository.Merchant, error)
}

type CreateMerchantRequest struct {
	Name          string  `json:"name"`
	LegalName     *string `json:"legal_name,omitempty"`
	Email         string  `json:"email"`
	Phone         *string `json:"phone,omitempty"`
	Website       *string `json:"website,omitempty"`
	Country       *string `json:"country,omitempty"`
	TaxID         *string `json:"tax_id,omitempty"`
	FeePercentage float64 `json:"fee_percentage"`
	WebhookURL    *string `json:"webhook_url,omitempty"`
	WebhookSecret *string `json:"webhook_secret,omitempty"`
}

type UpdateMerchantRequest struct {
	Name          *string `json:"name,omitempty"`
	LegalName     *string `json:"legal_name,omitempty"`
	Email         *string `json:"email,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Website       *string `json:"website,omitempty"`
	Country       *string `json:"country,omitempty"`
	TaxID         *string `json:"tax_id,omitempty"`
	Status        *string `json:"status,omitempty"`
	FeePercentage *float64 `json:"fee_percentage,omitempty"`
	WebhookURL    *string `json:"webhook_url,omitempty"`
	WebhookSecret *string `json:"webhook_secret,omitempty"`
}

type MerchantResponse struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	LegalName     *string    `json:"legal_name,omitempty"`
	Email         string     `json:"email"`
	Phone         *string    `json:"phone,omitempty"`
	Website       *string    `json:"website,omitempty"`
	Country       *string    `json:"country,omitempty"`
	TaxID         *string    `json:"tax_id,omitempty"`
	Status        string     `json:"status"`
	FeePercentage float64    `json:"fee_percentage"`
	WebhookURL    *string    `json:"webhook_url,omitempty"`
	WebhookSecret *string    `json:"webhook_secret,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type MerchantService struct {
	repo MerchantRepository
}

func NewMerchantService(repo MerchantRepository) *MerchantService {
	return &MerchantService{repo: repo}
}

func (s *MerchantService) CreateMerchant(ctx context.Context, req CreateMerchantRequest) (*MerchantResponse, error) {
	if req.Name == "" {
		return nil, ErrInvalidName
	}
	if req.Email == "" {
		return nil, ErrInvalidEmail
	}
	if req.FeePercentage == 0 {
		req.FeePercentage = 0.50
	}

	m := &repository.Merchant{
		Name:          req.Name,
		LegalName:     req.LegalName,
		Email:         req.Email,
		Phone:         req.Phone,
		Website:       req.Website,
		Country:       req.Country,
		TaxID:         req.TaxID,
		Status:        "active",
		FeePercentage: req.FeePercentage,
		WebhookURL:    req.WebhookURL,
		WebhookSecret: req.WebhookSecret,
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}

	return toResponse(m), nil
}

func (s *MerchantService) GetMerchant(ctx context.Context, id uuid.UUID) (*MerchantResponse, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrMerchantNotFound
	}
	return toResponse(m), nil
}

func (s *MerchantService) UpdateMerchant(ctx context.Context, id uuid.UUID, req UpdateMerchantRequest) (*MerchantResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrMerchantNotFound
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.LegalName != nil {
		existing.LegalName = req.LegalName
	}
	if req.Email != nil {
		existing.Email = *req.Email
	}
	if req.Phone != nil {
		existing.Phone = req.Phone
	}
	if req.Website != nil {
		existing.Website = req.Website
	}
	if req.Country != nil {
		existing.Country = req.Country
	}
	if req.TaxID != nil {
		existing.TaxID = req.TaxID
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.FeePercentage != nil {
		existing.FeePercentage = *req.FeePercentage
	}
	if req.WebhookURL != nil {
		existing.WebhookURL = req.WebhookURL
	}
	if req.WebhookSecret != nil {
		existing.WebhookSecret = req.WebhookSecret
	}

	if existing.Name == "" {
		return nil, ErrInvalidName
	}
	if existing.Email == "" {
		return nil, ErrInvalidEmail
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return toResponse(existing), nil
}

func (s *MerchantService) DeleteMerchant(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrMerchantNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *MerchantService) ListMerchants(ctx context.Context, page, pageSize int) ([]*MerchantResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	merchants, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*MerchantResponse, len(merchants))
	for i, m := range merchants {
		responses[i] = toResponse(m)
	}
	return responses, nil
}

func (s *MerchantService) SearchMerchants(ctx context.Context, query string, page, pageSize int) ([]*MerchantResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	merchants, err := s.repo.Search(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*MerchantResponse, len(merchants))
	for i, m := range merchants {
		responses[i] = toResponse(m)
	}
	return responses, nil
}

func toResponse(m *repository.Merchant) *MerchantResponse {
	return &MerchantResponse{
		ID:            m.ID,
		Name:          m.Name,
		LegalName:     m.LegalName,
		Email:         m.Email,
		Phone:         m.Phone,
		Website:       m.Website,
		Country:       m.Country,
		TaxID:         m.TaxID,
		Status:        m.Status,
		FeePercentage: m.FeePercentage,
		WebhookURL:    m.WebhookURL,
		WebhookSecret: m.WebhookSecret,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
