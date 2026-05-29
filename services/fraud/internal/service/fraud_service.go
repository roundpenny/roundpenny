package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/fraud/internal/repository"
)

var (
	ErrRuleNotFound      = errors.New("fraud rule not found")
	ErrAlertNotFound     = errors.New("fraud alert not found")
	ErrInvalidRuleName   = errors.New("rule name is required")
	ErrInvalidRuleType   = errors.New("rule type is required")
	ErrInvalidSeverity   = errors.New("severity must be low, medium, high, or critical")
	ErrInvalidStatus     = errors.New("status must be open, investigated, resolved, or false_positive")
)

type FraudRuleRepository interface {
	CreateRule(ctx context.Context, rule *repository.FraudRule) error
	GetRule(ctx context.Context, id uuid.UUID) (*repository.FraudRule, error)
	ListRules(ctx context.Context, limit, offset int) ([]*repository.FraudRule, error)
	UpdateRule(ctx context.Context, rule *repository.FraudRule) error
}

type FraudAlertRepository interface {
	CreateAlert(ctx context.Context, alert *repository.FraudAlert) error
	GetAlert(ctx context.Context, id uuid.UUID) (*repository.FraudAlert, error)
	ListAlerts(ctx context.Context, limit, offset int) ([]*repository.FraudAlert, error)
	UpdateAlertStatus(ctx context.Context, id uuid.UUID, status string) error
	GetAlertsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repository.FraudAlert, error)
	GetAlertsBySeverity(ctx context.Context, severity string, limit, offset int) ([]*repository.FraudAlert, error)
}

type CreateRuleRequest struct {
	Name     string          `json:"name"`
	RuleType string          `json:"rule_type"`
	Config   json.RawMessage `json:"config"`
	IsActive *bool           `json:"is_active,omitempty"`
	Priority *int            `json:"priority,omitempty"`
}

type UpdateRuleRequest struct {
	Name     *string          `json:"name,omitempty"`
	RuleType *string          `json:"rule_type,omitempty"`
	Config   *json.RawMessage `json:"config,omitempty"`
	IsActive *bool            `json:"is_active,omitempty"`
	Priority *int             `json:"priority,omitempty"`
}

type RuleResponse struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	RuleType  string          `json:"rule_type"`
	Config    json.RawMessage `json:"config"`
	IsActive  bool            `json:"is_active"`
	Priority  int             `json:"priority"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type CreateAlertRequest struct {
	UserID        uuid.UUID `json:"user_id"`
	RuleID        uuid.UUID `json:"rule_id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	Severity      string    `json:"severity"`
	Description   *string   `json:"description,omitempty"`
}

type AlertResponse struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	RuleID        uuid.UUID  `json:"rule_id"`
	TransactionID uuid.UUID  `json:"transaction_id"`
	Severity      string     `json:"severity"`
	Status        string     `json:"status"`
	Description   *string    `json:"description,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
}

type FraudService struct {
	ruleRepo  FraudRuleRepository
	alertRepo FraudAlertRepository
}

func NewFraudService(ruleRepo FraudRuleRepository, alertRepo FraudAlertRepository) *FraudService {
	return &FraudService{ruleRepo: ruleRepo, alertRepo: alertRepo}
}

func (s *FraudService) CreateRule(ctx context.Context, req CreateRuleRequest) (*RuleResponse, error) {
	if req.Name == "" {
		return nil, ErrInvalidRuleName
	}
	if req.RuleType == "" {
		return nil, ErrInvalidRuleType
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	priority := 0
	if req.Priority != nil {
		priority = *req.Priority
	}
	if req.Config == nil {
		req.Config = json.RawMessage(`{}`)
	}

	rule := &repository.FraudRule{
		Name:     req.Name,
		RuleType: req.RuleType,
		Config:   req.Config,
		IsActive: isActive,
		Priority: priority,
	}

	if err := s.ruleRepo.CreateRule(ctx, rule); err != nil {
		return nil, err
	}

	return ruleToResponse(rule), nil
}

func (s *FraudService) GetRule(ctx context.Context, id uuid.UUID) (*RuleResponse, error) {
	rule, err := s.ruleRepo.GetRule(ctx, id)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, ErrRuleNotFound
	}
	return ruleToResponse(rule), nil
}

func (s *FraudService) ListRules(ctx context.Context, page, pageSize int) ([]*RuleResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rules, err := s.ruleRepo.ListRules(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*RuleResponse, len(rules))
	for i, r := range rules {
		responses[i] = ruleToResponse(r)
	}
	return responses, nil
}

func (s *FraudService) UpdateRule(ctx context.Context, id uuid.UUID, req UpdateRuleRequest) (*RuleResponse, error) {
	existing, err := s.ruleRepo.GetRule(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrRuleNotFound
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.RuleType != nil {
		existing.RuleType = *req.RuleType
	}
	if req.Config != nil {
		existing.Config = *req.Config
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if req.Priority != nil {
		existing.Priority = *req.Priority
	}

	if existing.Name == "" {
		return nil, ErrInvalidRuleName
	}
	if existing.RuleType == "" {
		return nil, ErrInvalidRuleType
	}

	if err := s.ruleRepo.UpdateRule(ctx, existing); err != nil {
		return nil, err
	}

	return ruleToResponse(existing), nil
}

func (s *FraudService) CreateAlert(ctx context.Context, req CreateAlertRequest) (*AlertResponse, error) {
	if req.Severity != "low" && req.Severity != "medium" && req.Severity != "high" && req.Severity != "critical" {
		return nil, ErrInvalidSeverity
	}

	alert := &repository.FraudAlert{
		UserID:        req.UserID,
		RuleID:        req.RuleID,
		TransactionID: req.TransactionID,
		Severity:      req.Severity,
		Status:        "open",
		Description:   req.Description,
	}

	if err := s.alertRepo.CreateAlert(ctx, alert); err != nil {
		return nil, err
	}

	return alertToResponse(alert), nil
}

func (s *FraudService) GetAlert(ctx context.Context, id uuid.UUID) (*AlertResponse, error) {
	alert, err := s.alertRepo.GetAlert(ctx, id)
	if err != nil {
		return nil, err
	}
	if alert == nil {
		return nil, ErrAlertNotFound
	}
	return alertToResponse(alert), nil
}

func (s *FraudService) ListAlerts(ctx context.Context, page, pageSize int) ([]*AlertResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	alerts, err := s.alertRepo.ListAlerts(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*AlertResponse, len(alerts))
	for i, a := range alerts {
		responses[i] = alertToResponse(a)
	}
	return responses, nil
}

func (s *FraudService) UpdateAlertStatus(ctx context.Context, id uuid.UUID, status string) (*AlertResponse, error) {
	if status != "open" && status != "investigated" && status != "resolved" && status != "false_positive" {
		return nil, ErrInvalidStatus
	}

	existing, err := s.alertRepo.GetAlert(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrAlertNotFound
	}

	if err := s.alertRepo.UpdateAlertStatus(ctx, id, status); err != nil {
		return nil, err
	}

	existing.Status = status
	if status == "resolved" || status == "false_positive" {
		now := time.Now()
		existing.ResolvedAt = &now
	}

	return alertToResponse(existing), nil
}

func (s *FraudService) GetAlertsByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*AlertResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	alerts, err := s.alertRepo.GetAlertsByUser(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*AlertResponse, len(alerts))
	for i, a := range alerts {
		responses[i] = alertToResponse(a)
	}
	return responses, nil
}

func (s *FraudService) GetAlertsBySeverity(ctx context.Context, severity string, page, pageSize int) ([]*AlertResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	alerts, err := s.alertRepo.GetAlertsBySeverity(ctx, severity, pageSize, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*AlertResponse, len(alerts))
	for i, a := range alerts {
		responses[i] = alertToResponse(a)
	}
	return responses, nil
}

func ruleToResponse(r *repository.FraudRule) *RuleResponse {
	return &RuleResponse{
		ID:        r.ID,
		Name:      r.Name,
		RuleType:  r.RuleType,
		Config:    r.Config,
		IsActive:  r.IsActive,
		Priority:  r.Priority,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func alertToResponse(a *repository.FraudAlert) *AlertResponse {
	return &AlertResponse{
		ID:            a.ID,
		UserID:        a.UserID,
		RuleID:        a.RuleID,
		TransactionID: a.TransactionID,
		Severity:      a.Severity,
		Status:        a.Status,
		Description:   a.Description,
		CreatedAt:     a.CreatedAt,
		ResolvedAt:    a.ResolvedAt,
	}
}
