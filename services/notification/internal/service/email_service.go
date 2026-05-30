package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/roundup-platform/pkg/email"
	"github.com/roundup-platform/services/notification/internal/repository"
)

type EmailService struct {
	repo *repository.EmailRepository
	mail *email.Client
}

func NewEmailService(repo *repository.EmailRepository, mail *email.Client) *EmailService {
	return &EmailService{repo: repo, mail: mail}
}

func (s *EmailService) Send(ctx context.Context, userID, toEmail, subject, body, html string) (*repository.EmailLog, error) {
	err := s.mail.Send(email.SendEmailParams{
		To:      toEmail,
		Subject: subject,
		Body:    body,
		HTML:    html,
	})

	status := "sent"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
		slog.Error("email send failed", "to", toEmail, "error", err)
	}

	log, dbErr := s.repo.Insert(ctx, userID, toEmail, subject, body, html, status, errMsg)
	if dbErr != nil {
		return nil, fmt.Errorf("db insert error: %w", dbErr)
	}

	if err != nil {
		return log, err
	}

	return log, nil
}

func (s *EmailService) List(ctx context.Context, page, pageSize int) ([]repository.EmailLog, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, pageSize, offset)
}
