// Copyright (c) 2026 RoundPenny. All rights reserved.

package email

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Client struct {
	apiKey string
	from   string
	client *http.Client
	mock   bool
}

type SendEmailParams struct {
	To      string
	Subject string
	Body    string
	HTML    string
}

type sendGridRequest struct {
	Personalizations []sendGridPersonalization `json:"personalizations"`
	From             sendGridEmail             `json:"from"`
	Subject          string                    `json:"subject"`
	Content          []sendGridContent         `json:"content"`
}

type sendGridPersonalization struct {
	To []sendGridEmail `json:"to"`
}

type sendGridEmail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type sendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

var (
	ErrMissingAPIKey = errors.New("SENDGRID_API_KEY not set")
	ErrSendFailed    = errors.New("email send failed")
)

func NewClient() *Client {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		from = "noreply@roundpenny.com"
	}

	if apiKey == "" {
		slog.Warn("SENDGRID_API_KEY not set, using mock email mode")
	}

	mock := apiKey == ""

	return &Client{
		apiKey: apiKey,
		from:   from,
		client: &http.Client{Timeout: 10 * time.Second},
		mock:   mock,
	}
}

func (c *Client) Send(params SendEmailParams) error {
	if c.mock {
		return c.mockSend(params)
	}
	return c.sendGridSend(params)
}

func (c *Client) mockSend(params SendEmailParams) error {
	slog.Info("mock email",
		"to", params.To,
		"subject", params.Subject,
		"body_len", len(params.Body),
		"html_len", len(params.HTML),
	)
	return nil
}

func (c *Client) sendGridSend(params SendEmailParams) error {
	contentType := "text/plain"
	value := params.Body
	if params.HTML != "" {
		contentType = "text/html"
		value = params.HTML
	}

	body := sendGridRequest{
		Personalizations: []sendGridPersonalization{
			{To: []sendGridEmail{{Email: params.To}}},
		},
		From:    sendGridEmail{Email: c.from, Name: "RoundPenny"},
		Subject: params.Subject,
		Content: []sendGridContent{
			{Type: contentType, Value: value},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("%w: status %d", ErrSendFailed, resp.StatusCode)
	}

	slog.Info("email sent", "to", params.To, "subject", params.Subject)
	return nil
}
