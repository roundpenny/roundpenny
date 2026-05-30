package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailLog struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id,omitempty"`
	ToEmail   string     `json:"to_email"`
	Subject   string     `json:"subject"`
	Body      string     `json:"body,omitempty"`
	HTML      string     `json:"html,omitempty"`
	Status    string     `json:"status"`
	Error     string     `json:"error,omitempty"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type EmailRepository struct {
	pool *pgxpool.Pool
}

func NewEmailRepository(pool *pgxpool.Pool) *EmailRepository {
	return &EmailRepository{pool: pool}
}

func (r *EmailRepository) Insert(ctx context.Context, userID, toEmail, subject, body, html, status, errMsg string) (*EmailLog, error) {
	var log EmailLog
	err := r.pool.QueryRow(ctx,
		`INSERT INTO email_logs (user_id, to_email, subject, body, html, status, error, sent_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, CASE WHEN $6 = 'sent' THEN NOW() ELSE NULL END)
		 RETURNING id, user_id, to_email, subject, body, html, status, error, sent_at, created_at`,
		userID, toEmail, subject, body, html, status, errMsg,
	).Scan(&log.ID, &log.UserID, &log.ToEmail, &log.Subject, &log.Body, &log.HTML, &log.Status, &log.Error, &log.SentAt, &log.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *EmailRepository) List(ctx context.Context, limit, offset int) ([]EmailLog, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM email_logs`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, to_email, subject, body, html, status, error, sent_at, created_at
		 FROM email_logs ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []EmailLog
	for rows.Next() {
		var l EmailLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.ToEmail, &l.Subject, &l.Body, &l.HTML, &l.Status, &l.Error, &l.SentAt, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	return logs, total, nil
}
