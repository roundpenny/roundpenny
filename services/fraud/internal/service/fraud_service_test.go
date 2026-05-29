package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/roundup-platform/services/fraud/internal/repository"
)

type mockFraudRuleRepo struct {
	createFn func(ctx context.Context, rule *repository.FraudRule) error
	getByIDFn func(ctx context.Context, id uuid.UUID) (*repository.FraudRule, error)
	listFn    func(ctx context.Context, limit, offset int) ([]*repository.FraudRule, error)
	updateFn  func(ctx context.Context, rule *repository.FraudRule) error
}

func (m *mockFraudRuleRepo) CreateRule(ctx context.Context, rule *repository.FraudRule) error {
	return m.createFn(ctx, rule)
}
func (m *mockFraudRuleRepo) GetRule(ctx context.Context, id uuid.UUID) (*repository.FraudRule, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockFraudRuleRepo) ListRules(ctx context.Context, limit, offset int) ([]*repository.FraudRule, error) {
	return m.listFn(ctx, limit, offset)
}
func (m *mockFraudRuleRepo) UpdateRule(ctx context.Context, rule *repository.FraudRule) error {
	return m.updateFn(ctx, rule)
}

type mockFraudAlertRepo struct {
	createFn      func(ctx context.Context, alert *repository.FraudAlert) error
	getByIDFn     func(ctx context.Context, id uuid.UUID) (*repository.FraudAlert, error)
	listFn        func(ctx context.Context, limit, offset int) ([]*repository.FraudAlert, error)
	updateStatusFn func(ctx context.Context, id uuid.UUID, status string) error
	getByUserFn   func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repository.FraudAlert, error)
	getBySeverityFn func(ctx context.Context, severity string, limit, offset int) ([]*repository.FraudAlert, error)
}

func (m *mockFraudAlertRepo) CreateAlert(ctx context.Context, alert *repository.FraudAlert) error {
	return m.createFn(ctx, alert)
}
func (m *mockFraudAlertRepo) GetAlert(ctx context.Context, id uuid.UUID) (*repository.FraudAlert, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockFraudAlertRepo) ListAlerts(ctx context.Context, limit, offset int) ([]*repository.FraudAlert, error) {
	return m.listFn(ctx, limit, offset)
}
func (m *mockFraudAlertRepo) UpdateAlertStatus(ctx context.Context, id uuid.UUID, status string) error {
	return m.updateStatusFn(ctx, id, status)
}
func (m *mockFraudAlertRepo) GetAlertsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repository.FraudAlert, error) {
	return m.getByUserFn(ctx, userID, limit, offset)
}
func (m *mockFraudAlertRepo) GetAlertsBySeverity(ctx context.Context, severity string, limit, offset int) ([]*repository.FraudAlert, error) {
	return m.getBySeverityFn(ctx, severity, limit, offset)
}

func TestCreateRule_Success(t *testing.T) {
	ruleID := uuid.New()
	now := time.Now()

	mockRule := &mockFraudRuleRepo{
		createFn: func(ctx context.Context, rule *repository.FraudRule) error {
			rule.ID = ruleID
			rule.CreatedAt = now
			rule.UpdatedAt = now
			return nil
		},
	}

	svc := NewFraudService(mockRule, &mockFraudAlertRepo{})
	resp, err := svc.CreateRule(context.Background(), CreateRuleRequest{
		Name:     "High Amount Rule",
		RuleType: "amount",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Name != "High Amount Rule" {
		t.Errorf("expected High Amount Rule, got %s", resp.Name)
	}
	if resp.RuleType != "amount" {
		t.Errorf("expected amount, got %s", resp.RuleType)
	}
	if !resp.IsActive {
		t.Errorf("expected is_active true")
	}
}

func TestCreateRule_InvalidName(t *testing.T) {
	svc := NewFraudService(&mockFraudRuleRepo{}, &mockFraudAlertRepo{})
	_, err := svc.CreateRule(context.Background(), CreateRuleRequest{
		RuleType: "amount",
	})
	if err != ErrInvalidRuleName {
		t.Errorf("expected ErrInvalidRuleName, got %v", err)
	}
}

func TestCreateRule_InvalidType(t *testing.T) {
	svc := NewFraudService(&mockFraudRuleRepo{}, &mockFraudAlertRepo{})
	_, err := svc.CreateRule(context.Background(), CreateRuleRequest{
		Name: "Test Rule",
	})
	if err != ErrInvalidRuleType {
		t.Errorf("expected ErrInvalidRuleType, got %v", err)
	}
}

func TestGetRule_NotFound(t *testing.T) {
	mock := &mockFraudRuleRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.FraudRule, error) {
			return nil, nil
		},
	}

	svc := NewFraudService(mock, &mockFraudAlertRepo{})
	_, err := svc.GetRule(context.Background(), uuid.New())
	if err != ErrRuleNotFound {
		t.Errorf("expected ErrRuleNotFound, got %v", err)
	}
}

func TestGetRule_Success(t *testing.T) {
	ruleID := uuid.New()

	mock := &mockFraudRuleRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.FraudRule, error) {
			return &repository.FraudRule{
				ID:       ruleID,
				Name:     "Test Rule",
				RuleType: "amount",
				Config:   json.RawMessage(`{}`),
				IsActive: true,
				Priority: 10,
			}, nil
		},
	}

	svc := NewFraudService(mock, &mockFraudAlertRepo{})
	resp, err := svc.GetRule(context.Background(), ruleID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != ruleID {
		t.Errorf("expected %v, got %v", ruleID, resp.ID)
	}
	if resp.Priority != 10 {
		t.Errorf("expected priority 10, got %d", resp.Priority)
	}
}

func TestUpdateRule_Success(t *testing.T) {
	ruleID := uuid.New()

	mock := &mockFraudRuleRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.FraudRule, error) {
			return &repository.FraudRule{
				ID:       ruleID,
				Name:     "Old Rule",
				RuleType: "amount",
				Config:   json.RawMessage(`{}`),
				IsActive: true,
				Priority: 5,
			}, nil
		},
		updateFn: func(ctx context.Context, rule *repository.FraudRule) error {
			return nil
		},
	}

	newName := "Updated Rule"
	svc := NewFraudService(mock, &mockFraudAlertRepo{})
	resp, err := svc.UpdateRule(context.Background(), ruleID, UpdateRuleRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Name != "Updated Rule" {
		t.Errorf("expected Updated Rule, got %s", resp.Name)
	}
}

func TestUpdateRule_NotFound(t *testing.T) {
	mock := &mockFraudRuleRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.FraudRule, error) {
			return nil, nil
		},
	}

	svc := NewFraudService(mock, &mockFraudAlertRepo{})
	_, err := svc.UpdateRule(context.Background(), uuid.New(), UpdateRuleRequest{})
	if err != ErrRuleNotFound {
		t.Errorf("expected ErrRuleNotFound, got %v", err)
	}
}

func TestListRules_DefaultPagination(t *testing.T) {
	mock := &mockFraudRuleRepo{
		listFn: func(ctx context.Context, limit, offset int) ([]*repository.FraudRule, error) {
			if limit != 20 || offset != 0 {
				t.Errorf("expected limit=20, offset=0, got limit=%d offset=%d", limit, offset)
			}
			return []*repository.FraudRule{}, nil
		},
	}

	svc := NewFraudService(mock, &mockFraudAlertRepo{})
	resp, err := svc.ListRules(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list")
	}
}

func TestCreateAlert_Success(t *testing.T) {
	alertID := uuid.New()
	now := time.Now()

	mockAlert := &mockFraudAlertRepo{
		createFn: func(ctx context.Context, alert *repository.FraudAlert) error {
			alert.ID = alertID
			alert.CreatedAt = now
			return nil
		},
	}

	svc := NewFraudService(&mockFraudRuleRepo{}, mockAlert)
	resp, err := svc.CreateAlert(context.Background(), CreateAlertRequest{
		UserID:        uuid.New(),
		RuleID:        uuid.New(),
		TransactionID: uuid.New(),
		Severity:      "high",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != alertID {
		t.Errorf("expected %v, got %v", alertID, resp.ID)
	}
	if resp.Status != "open" {
		t.Errorf("expected open, got %s", resp.Status)
	}
	if resp.Severity != "high" {
		t.Errorf("expected high, got %s", resp.Severity)
	}
}

func TestCreateAlert_InvalidSeverity(t *testing.T) {
	svc := NewFraudService(&mockFraudRuleRepo{}, &mockFraudAlertRepo{})
	_, err := svc.CreateAlert(context.Background(), CreateAlertRequest{
		UserID:        uuid.New(),
		RuleID:        uuid.New(),
		TransactionID: uuid.New(),
		Severity:      "unknown",
	})
	if err != ErrInvalidSeverity {
		t.Errorf("expected ErrInvalidSeverity, got %v", err)
	}
}

func TestGetAlert_NotFound(t *testing.T) {
	mock := &mockFraudAlertRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.FraudAlert, error) {
			return nil, nil
		},
	}

	svc := NewFraudService(&mockFraudRuleRepo{}, mock)
	_, err := svc.GetAlert(context.Background(), uuid.New())
	if err != ErrAlertNotFound {
		t.Errorf("expected ErrAlertNotFound, got %v", err)
	}
}

func TestGetAlert_Success(t *testing.T) {
	alertID := uuid.New()

	mock := &mockFraudAlertRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.FraudAlert, error) {
			return &repository.FraudAlert{
				ID:            alertID,
				UserID:        uuid.New(),
				RuleID:        uuid.New(),
				TransactionID: uuid.New(),
				Severity:      "critical",
				Status:        "open",
			}, nil
		},
	}

	svc := NewFraudService(&mockFraudRuleRepo{}, mock)
	resp, err := svc.GetAlert(context.Background(), alertID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != alertID {
		t.Errorf("expected %v, got %v", alertID, resp.ID)
	}
	if resp.Severity != "critical" {
		t.Errorf("expected critical, got %s", resp.Severity)
	}
}

func TestUpdateAlertStatus_Success(t *testing.T) {
	alertID := uuid.New()

	mock := &mockFraudAlertRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.FraudAlert, error) {
			return &repository.FraudAlert{
				ID:            alertID,
				UserID:        uuid.New(),
				RuleID:        uuid.New(),
				TransactionID: uuid.New(),
				Severity:      "medium",
				Status:        "open",
			}, nil
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status string) error {
			return nil
		},
	}

	svc := NewFraudService(&mockFraudRuleRepo{}, mock)
	resp, err := svc.UpdateAlertStatus(context.Background(), alertID, "resolved")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "resolved" {
		t.Errorf("expected resolved, got %s", resp.Status)
	}
	if resp.ResolvedAt == nil {
		t.Errorf("expected resolved_at to be set")
	}
}

func TestUpdateAlertStatus_NotFound(t *testing.T) {
	mock := &mockFraudAlertRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.FraudAlert, error) {
			return nil, nil
		},
	}

	svc := NewFraudService(&mockFraudRuleRepo{}, mock)
	_, err := svc.UpdateAlertStatus(context.Background(), uuid.New(), "resolved")
	if err != ErrAlertNotFound {
		t.Errorf("expected ErrAlertNotFound, got %v", err)
	}
}

func TestUpdateAlertStatus_InvalidStatus(t *testing.T) {
	mock := &mockFraudAlertRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*repository.FraudAlert, error) {
			return &repository.FraudAlert{ID: id, Status: "open"}, nil
		},
	}

	svc := NewFraudService(&mockFraudRuleRepo{}, mock)
	_, err := svc.UpdateAlertStatus(context.Background(), uuid.New(), "invalid")
	if err != ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestGetAlertsByUser_Success(t *testing.T) {
	userID := uuid.New()
	alertID := uuid.New()

	mock := &mockFraudAlertRepo{
		getByUserFn: func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*repository.FraudAlert, error) {
			if uid != userID {
				t.Errorf("expected userID %v, got %v", userID, uid)
			}
			return []*repository.FraudAlert{
				{ID: alertID, UserID: userID, RuleID: uuid.New(), TransactionID: uuid.New(), Severity: "low", Status: "open"},
			}, nil
		},
	}

	svc := NewFraudService(&mockFraudRuleRepo{}, mock)
	resp, err := svc.GetAlertsByUser(context.Background(), userID, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 alert, got %d", len(resp))
	}
	if resp[0].UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, resp[0].UserID)
	}
}

func TestGetAlertsBySeverity_Success(t *testing.T) {
	alertID := uuid.New()

	mock := &mockFraudAlertRepo{
		getBySeverityFn: func(ctx context.Context, severity string, limit, offset int) ([]*repository.FraudAlert, error) {
			if severity != "critical" {
				t.Errorf("expected critical, got %s", severity)
			}
			return []*repository.FraudAlert{
				{ID: alertID, UserID: uuid.New(), RuleID: uuid.New(), TransactionID: uuid.New(), Severity: "critical", Status: "open"},
			}, nil
		},
	}

	svc := NewFraudService(&mockFraudRuleRepo{}, mock)
	resp, err := svc.GetAlertsBySeverity(context.Background(), "critical", 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 alert, got %d", len(resp))
	}
	if resp[0].Severity != "critical" {
		t.Errorf("expected critical, got %s", resp[0].Severity)
	}
}

func TestListAlerts_DefaultPagination(t *testing.T) {
	mock := &mockFraudAlertRepo{
		listFn: func(ctx context.Context, limit, offset int) ([]*repository.FraudAlert, error) {
			if limit != 20 || offset != 0 {
				t.Errorf("expected limit=20, offset=0, got limit=%d offset=%d", limit, offset)
			}
			return []*repository.FraudAlert{}, nil
		},
	}

	svc := NewFraudService(&mockFraudRuleRepo{}, mock)
	resp, err := svc.ListAlerts(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list")
	}
}
