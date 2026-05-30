// Copyright (c) 2026 RoundPenny. All rights reserved.

package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func TestNewAuditLogger_dev_mode(t *testing.T) {
	l := NewAuditLogger(false, "")
	if l.writer == nil {
		t.Fatal("expected non-nil writer")
	}
	if l.encoder == nil {
		t.Fatal("expected non-nil encoder")
	}
}

func TestNewAuditLogger_prod_mode_without_kafka(t *testing.T) {
	l := NewAuditLogger(true, "")
	if l.writer == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestAuditLogger_log_with_metadata(t *testing.T) {
	var buf bytes.Buffer
	l := &AuditLogger{
		prodMode: true,
		writer:   &buf,
		encoder:  json.NewEncoder(&buf),
	}

	l.Log(context.Background(), ActionRegister, "user", "user_123", "user_123", map[string]any{
		"ip": "192.168.1.1",
	})

	var entry AuditEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entry.Action != ActionRegister {
		t.Fatalf("got %q, want %q", entry.Action, ActionRegister)
	}
	if entry.ResourceType != "user" {
		t.Fatalf("got %q", entry.ResourceType)
	}
	if entry.ResourceID != "user_123" {
		t.Fatalf("got %q", entry.ResourceID)
	}
	if entry.UserID != "user_123" {
		t.Fatalf("got %q", entry.UserID)
	}
	if entry.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if entry.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
	if entry.Metadata["ip"] != "192.168.1.1" {
		t.Fatalf("got %v", entry.Metadata["ip"])
	}
}

func TestAuditLogger_log_with_context_values(t *testing.T) {
	var buf bytes.Buffer
	l := &AuditLogger{
		prodMode: true,
		writer:   &buf,
		encoder:  json.NewEncoder(&buf),
	}

	ctx := context.WithValue(context.Background(), "request_id", "req_abc")
	ctx = context.WithValue(ctx, "client_ip", "10.0.0.1")
	ctx = context.WithValue(ctx, "user_agent", "test-agent/1.0")

	l.Log(ctx, ActionLogin, "", "", "", nil)

	var entry AuditEntry
	json.Unmarshal(buf.Bytes(), &entry)

	if entry.RequestID != "req_abc" {
		t.Fatalf("got %q", entry.RequestID)
	}
	if entry.ClientIP != "10.0.0.1" {
		t.Fatalf("got %q", entry.ClientIP)
	}
	if entry.UserAgent != "test-agent/1.0" {
		t.Fatalf("got %q", entry.UserAgent)
	}
}

func TestAuditLogger_log_in_dev_mode(t *testing.T) {
	var buf bytes.Buffer
	l := &AuditLogger{
		prodMode: false,
		writer:   &buf,
		encoder:  json.NewEncoder(&buf),
	}

	l.Log(context.Background(), ActionKYCSubmit, "kyc", "kyc_123", "user_123", nil)
}

func TestKafkaWriter(t *testing.T) {
	w := newKafkaWriter("broker:9092")
	n, err := w.Write([]byte(`{"test": "data"}`))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 16 {
		t.Fatalf("got %d, want 16", n)
	}
}

func TestAuditEntry_actions(t *testing.T) {
	actions := []Action{
		ActionRegister, ActionLogin, ActionLogout, ActionRefresh,
		ActionMFAConfigure, ActionMFAVerify, ActionMFAEnable, ActionMFADisable,
		ActionEmailVerify, ActionKYCSubmit, ActionKYCStatus,
		ActionPasswordChange, ActionProfileUpdate, ActionPreferencesUpdate,
		ActionMerchantCreate, ActionMerchantUpdate, ActionMerchantDelete,
		ActionPaymentCreate, ActionPaymentRefund,
		ActionWebhookCreate, ActionWebhookDelete,
		ActionAnalyticsEvent, ActionTransactionCreate,
		ActionRoundupTrigger, ActionInvestmentCreate, ActionInvestmentWithdraw,
	}
	for _, a := range actions {
		if a == "" {
			t.Fatal("action should not be empty")
		}
	}
}
