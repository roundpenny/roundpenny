// Copyright (c) 2026 RoundPenny. All rights reserved.

package audit

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Action string

const (
	ActionRegister            Action = "user.register"
	ActionLogin               Action = "user.login"
	ActionLogout              Action = "user.logout"
	ActionRefresh             Action = "user.refresh"
	ActionMFAConfigure        Action = "mfa.configure"
	ActionMFAVerify           Action = "mfa.verify"
	ActionMFAEnable           Action = "mfa.enable"
	ActionMFADisable          Action = "mfa.disable"
	ActionEmailVerify         Action = "email.verify"
	ActionKYCSubmit           Action = "kyc.submit"
	ActionKYCStatus           Action = "kyc.status"
	ActionPasswordChange      Action = "password.change"
	ActionProfileUpdate       Action = "profile.update"
	ActionPreferencesUpdate   Action = "preferences.update"
	ActionMerchantCreate      Action = "merchant.create"
	ActionMerchantUpdate      Action = "merchant.update"
	ActionMerchantDelete      Action = "merchant.delete"
	ActionPaymentCreate       Action = "payment.create"
	ActionPaymentRefund       Action = "payment.refund"
	ActionWebhookCreate       Action = "webhook.create"
	ActionWebhookDelete       Action = "webhook.delete"
	ActionAnalyticsEvent      Action = "analytics.event"
	ActionTransactionCreate   Action = "transaction.create"
	ActionRoundupTrigger      Action = "roundup.trigger"
	ActionInvestmentCreate    Action = "investment.create"
	ActionInvestmentWithdraw  Action = "investment.withdraw"
)

type AuditEntry struct {
	ID           string            `json:"id"`
	Timestamp    time.Time         `json:"timestamp"`
	Action       Action            `json:"action"`
	ResourceType string            `json:"resource_type,omitempty"`
	ResourceID   string            `json:"resource_id,omitempty"`
	UserID       string            `json:"user_id,omitempty"`
	Metadata     map[string]any    `json:"metadata,omitempty"`
	RequestID    string            `json:"request_id,omitempty"`
	ClientIP     string            `json:"client_ip,omitempty"`
	UserAgent    string            `json:"user_agent,omitempty"`
}

type AuditLogger struct {
	mu       sync.Mutex
	encoder  *json.Encoder
	writer   io.Writer
	prodMode bool
}

func NewAuditLogger(prodMode bool, kafkaBrokers string) *AuditLogger {
	logger := &AuditLogger{
		prodMode: prodMode,
	}

	if prodMode && kafkaBrokers != "" {
		logger.writer = newKafkaWriter(kafkaBrokers)
	} else {
		logger.writer = os.Stdout
	}

	logger.encoder = json.NewEncoder(logger.writer)
	return logger
}

func (a *AuditLogger) Log(ctx context.Context, action Action, resourceType, resourceID, userID string, metadata map[string]any) {
	entry := AuditEntry{
		ID:           uuid.New().String(),
		Timestamp:    time.Now().UTC(),
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserID:       userID,
		Metadata:     metadata,
	}

	if reqID, ok := ctx.Value("request_id").(string); ok {
		entry.RequestID = reqID
	}
	if clientIP, ok := ctx.Value("client_ip").(string); ok {
		entry.ClientIP = clientIP
	}
	if userAgent, ok := ctx.Value("user_agent").(string); ok {
		entry.UserAgent = userAgent
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.prodMode {
		if err := a.encoder.Encode(entry); err != nil {
			slog.Error("audit log write failed", "error", err)
		}
	} else {
		slog.Info("audit",
			"action", entry.Action,
			"resource_type", entry.ResourceType,
			"resource_id", entry.ResourceID,
			"user_id", entry.UserID,
			"request_id", entry.RequestID,
		)
	}
}

type kafkaAuditWriter struct {
	topic string
}

func newKafkaWriter(brokers string) io.Writer {
	return &kafkaAuditWriter{topic: "audit-logs"}
}

func (w *kafkaAuditWriter) Write(p []byte) (int, error) {
	slog.Info("audit kafka would send", "topic", w.topic, "message", string(p))
	return len(p), nil
}
