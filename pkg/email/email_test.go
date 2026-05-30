// Copyright (c) 2026 RoundPenny. All rights reserved.

package email

import (
	"testing"
)

func TestNewClient_mock(t *testing.T) {
	c := NewClient()
	if !c.mock {
		t.Fatal("expected mock mode when SENDGRID_API_KEY is not set")
	}
	if c.from != "noreply@roundpenny.com" {
		t.Fatalf("got %q, want %q", c.from, "noreply@roundpenny.com")
	}
}

func TestNewClient_custom_from(t *testing.T) {
	t.Setenv("EMAIL_FROM", "custom@example.com")
	c := NewClient()
	if c.from != "custom@example.com" {
		t.Fatalf("got %q, want %q", c.from, "custom@example.com")
	}
}

func TestSend_mock(t *testing.T) {
	c := &Client{mock: true}

	tests := []struct {
		name   string
		params SendEmailParams
	}{
		{"plain text", SendEmailParams{To: "user@test.com", Subject: "Hello", Body: "plain body"}},
		{"html", SendEmailParams{To: "user@test.com", Subject: "Hello", Body: "plain", HTML: "<html></html>"}},
		{"empty", SendEmailParams{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.Send(tt.params); err != nil {
				t.Fatalf("Send failed: %v", err)
			}
		})
	}
}

func TestMockSend(t *testing.T) {
	c := &Client{mock: true}
	if err := c.mockSend(SendEmailParams{To: "test@test.com", Subject: "Test", Body: "body"}); err != nil {
		t.Fatalf("mockSend failed: %v", err)
	}
}
