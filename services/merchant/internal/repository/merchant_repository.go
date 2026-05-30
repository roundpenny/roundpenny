// Copyright (c) 2026 RoundPenny. All rights reserved.

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/roundup-platform/pkg/db"
)

type Merchant struct {
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
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

type MerchantRepository struct {
	pool *db.Pool
}

func NewMerchantRepository(pool *db.Pool) *MerchantRepository {
	return &MerchantRepository{pool: pool}
}

func (r *MerchantRepository) Create(ctx context.Context, m *Merchant) error {
	query := `INSERT INTO merchants (name, legal_name, email, phone, website, country, tax_id, status, fee_percentage, webhook_url, webhook_secret)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	          RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		m.Name, m.LegalName, m.Email, m.Phone, m.Website, m.Country, m.TaxID,
		m.Status, m.FeePercentage, m.WebhookURL, m.WebhookSecret,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

func (r *MerchantRepository) GetByID(ctx context.Context, id uuid.UUID) (*Merchant, error) {
	query := `SELECT id, name, legal_name, email, phone, website, country, tax_id,
	                 status, fee_percentage, webhook_url, webhook_secret,
	                 created_at, updated_at, deleted_at
	          FROM merchants WHERE id = $1 AND deleted_at IS NULL`
	m := &Merchant{}
	var legalName, phone, website, country, taxID, webhookURL, webhookSecret *string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.Name, &legalName, &m.Email, &phone, &website, &country, &taxID,
		&m.Status, &m.FeePercentage, &webhookURL, &webhookSecret,
		&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if legalName != nil {
		m.LegalName = legalName
	}
	if phone != nil {
		m.Phone = phone
	}
	if website != nil {
		m.Website = website
	}
	if country != nil {
		m.Country = country
	}
	if taxID != nil {
		m.TaxID = taxID
	}
	if webhookURL != nil {
		m.WebhookURL = webhookURL
	}
	if webhookSecret != nil {
		m.WebhookSecret = webhookSecret
	}
	return m, nil
}

func (r *MerchantRepository) Update(ctx context.Context, m *Merchant) error {
	query := `UPDATE merchants SET name = $2, legal_name = $3, email = $4, phone = $5,
	          website = $6, country = $7, tax_id = $8, status = $9, fee_percentage = $10,
	          webhook_url = $11, webhook_secret = $12, updated_at = NOW()
	          WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query,
		m.ID, m.Name, m.LegalName, m.Email, m.Phone, m.Website, m.Country, m.TaxID,
		m.Status, m.FeePercentage, m.WebhookURL, m.WebhookSecret,
	)
	return err
}

func (r *MerchantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE merchants SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *MerchantRepository) Search(ctx context.Context, query string, limit, offset int) ([]*Merchant, error) {
	sql := `SELECT id, name, legal_name, email, phone, website, country, tax_id,
	               status, fee_percentage, webhook_url, webhook_secret,
	               created_at, updated_at, deleted_at
	        FROM merchants WHERE deleted_at IS NULL AND (name ILIKE $1 OR email ILIKE $1)
	        ORDER BY name ASC LIMIT $2 OFFSET $3`
	pattern := "%" + query + "%"
	rows, err := r.pool.Query(ctx, sql, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var merchants []*Merchant
	for rows.Next() {
		m := &Merchant{}
		var legalName, phone, website, country, taxID, webhookURL, webhookSecret *string
		if err := rows.Scan(
			&m.ID, &m.Name, &legalName, &m.Email, &phone, &website, &country, &taxID,
			&m.Status, &m.FeePercentage, &webhookURL, &webhookSecret,
			&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
		); err != nil {
			return nil, err
		}
		if legalName != nil {
			m.LegalName = legalName
		}
		if phone != nil {
			m.Phone = phone
		}
		if website != nil {
			m.Website = website
		}
		if country != nil {
			m.Country = country
		}
		if taxID != nil {
			m.TaxID = taxID
		}
		if webhookURL != nil {
			m.WebhookURL = webhookURL
		}
		if webhookSecret != nil {
			m.WebhookSecret = webhookSecret
		}
		merchants = append(merchants, m)
	}
	return merchants, nil
}

func (r *MerchantRepository) List(ctx context.Context, limit, offset int) ([]*Merchant, error) {
	sql := `SELECT id, name, legal_name, email, phone, website, country, tax_id,
	               status, fee_percentage, webhook_url, webhook_secret,
	               created_at, updated_at, deleted_at
	        FROM merchants WHERE deleted_at IS NULL
	        ORDER BY name ASC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, sql, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var merchants []*Merchant
	for rows.Next() {
		m := &Merchant{}
		var legalName, phone, website, country, taxID, webhookURL, webhookSecret *string
		if err := rows.Scan(
			&m.ID, &m.Name, &legalName, &m.Email, &phone, &website, &country, &taxID,
			&m.Status, &m.FeePercentage, &webhookURL, &webhookSecret,
			&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
		); err != nil {
			return nil, err
		}
		if legalName != nil {
			m.LegalName = legalName
		}
		if phone != nil {
			m.Phone = phone
		}
		if website != nil {
			m.Website = website
		}
		if country != nil {
			m.Country = country
		}
		if taxID != nil {
			m.TaxID = taxID
		}
		if webhookURL != nil {
			m.WebhookURL = webhookURL
		}
		if webhookSecret != nil {
			m.WebhookSecret = webhookSecret
		}
		merchants = append(merchants, m)
	}
	return merchants, nil
}
