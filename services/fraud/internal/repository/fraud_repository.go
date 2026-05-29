package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/roundup-platform/pkg/db"
)

type FraudRule struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	RuleType  string          `json:"rule_type"`
	Config    json.RawMessage `json:"config"`
	IsActive  bool            `json:"is_active"`
	Priority  int             `json:"priority"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type FraudAlert struct {
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

type FraudRepository struct {
	pool *db.Pool
}

func NewFraudRepository(pool *db.Pool) *FraudRepository {
	return &FraudRepository{pool: pool}
}

func (r *FraudRepository) CreateRule(ctx context.Context, rule *FraudRule) error {
	query := `INSERT INTO fraud_rules (name, rule_type, config, is_active, priority)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		rule.Name, rule.RuleType, rule.Config, rule.IsActive, rule.Priority,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)
}

func (r *FraudRepository) GetRule(ctx context.Context, id uuid.UUID) (*FraudRule, error) {
	query := `SELECT id, name, rule_type, config, is_active, priority, created_at, updated_at
	          FROM fraud_rules WHERE id = $1`
	rule := &FraudRule{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rule.ID, &rule.Name, &rule.RuleType, &rule.Config,
		&rule.IsActive, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return rule, nil
}

func (r *FraudRepository) ListRules(ctx context.Context, limit, offset int) ([]*FraudRule, error) {
	query := `SELECT id, name, rule_type, config, is_active, priority, created_at, updated_at
	          FROM fraud_rules ORDER BY priority DESC, name ASC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*FraudRule
	for rows.Next() {
		rule := &FraudRule{}
		if err := rows.Scan(
			&rule.ID, &rule.Name, &rule.RuleType, &rule.Config,
			&rule.IsActive, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *FraudRepository) UpdateRule(ctx context.Context, rule *FraudRule) error {
	query := `UPDATE fraud_rules SET name = $2, rule_type = $3, config = $4,
	          is_active = $5, priority = $6, updated_at = NOW()
	          WHERE id = $1`
	_, err := r.pool.Exec(ctx, query,
		rule.ID, rule.Name, rule.RuleType, rule.Config,
		rule.IsActive, rule.Priority,
	)
	return err
}

func (r *FraudRepository) CreateAlert(ctx context.Context, alert *FraudAlert) error {
	query := `INSERT INTO fraud_alerts (user_id, rule_id, transaction_id, severity, status, description)
	          VALUES ($1, $2, $3, $4, $5, $6)
	          RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query,
		alert.UserID, alert.RuleID, alert.TransactionID,
		alert.Severity, alert.Status, alert.Description,
	).Scan(&alert.ID, &alert.CreatedAt)
}

func (r *FraudRepository) GetAlert(ctx context.Context, id uuid.UUID) (*FraudAlert, error) {
	query := `SELECT id, user_id, rule_id, transaction_id, severity, status, description, created_at, resolved_at
	          FROM fraud_alerts WHERE id = $1`
	alert := &FraudAlert{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&alert.ID, &alert.UserID, &alert.RuleID, &alert.TransactionID,
		&alert.Severity, &alert.Status, &alert.Description,
		&alert.CreatedAt, &alert.ResolvedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return alert, nil
}

func (r *FraudRepository) ListAlerts(ctx context.Context, limit, offset int) ([]*FraudAlert, error) {
	query := `SELECT id, user_id, rule_id, transaction_id, severity, status, description, created_at, resolved_at
	          FROM fraud_alerts ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*FraudAlert
	for rows.Next() {
		alert := &FraudAlert{}
		if err := rows.Scan(
			&alert.ID, &alert.UserID, &alert.RuleID, &alert.TransactionID,
			&alert.Severity, &alert.Status, &alert.Description,
			&alert.CreatedAt, &alert.ResolvedAt,
		); err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	return alerts, nil
}

func (r *FraudRepository) UpdateAlertStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE fraud_alerts SET status = $2, resolved_at = CASE WHEN $2 IN ('resolved', 'false_positive') THEN NOW() ELSE resolved_at END
	          WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, status)
	return err
}

func (r *FraudRepository) GetAlertsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*FraudAlert, error) {
	query := `SELECT id, user_id, rule_id, transaction_id, severity, status, description, created_at, resolved_at
	          FROM fraud_alerts WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*FraudAlert
	for rows.Next() {
		alert := &FraudAlert{}
		if err := rows.Scan(
			&alert.ID, &alert.UserID, &alert.RuleID, &alert.TransactionID,
			&alert.Severity, &alert.Status, &alert.Description,
			&alert.CreatedAt, &alert.ResolvedAt,
		); err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	return alerts, nil
}

func (r *FraudRepository) GetAlertsBySeverity(ctx context.Context, severity string, limit, offset int) ([]*FraudAlert, error) {
	query := `SELECT id, user_id, rule_id, transaction_id, severity, status, description, created_at, resolved_at
	          FROM fraud_alerts WHERE severity = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, severity, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*FraudAlert
	for rows.Next() {
		alert := &FraudAlert{}
		if err := rows.Scan(
			&alert.ID, &alert.UserID, &alert.RuleID, &alert.TransactionID,
			&alert.Severity, &alert.Status, &alert.Description,
			&alert.CreatedAt, &alert.ResolvedAt,
		); err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	return alerts, nil
}
