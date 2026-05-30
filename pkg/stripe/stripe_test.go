// Copyright (c) 2026 RoundPenny. All rights reserved.

package stripe

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewClient_mock(t *testing.T) {
	c := NewClient()
	if c.client != nil {
		t.Fatal("expected nil client (mock mode) when STRIPE_API_KEY is not set")
	}
}

func TestCreatePaymentIntent_mock(t *testing.T) {
	c := &StripeClient{}

	params := CreatePaymentIntentParams{
		Amount:        1000,
		Currency:      "usd",
		Description:   "Test payment",
		PaymentMethod: "pm_card_visa",
		Metadata:      map[string]string{"order_id": "ord_123"},
	}

	pi, err := c.CreatePaymentIntent(params)
	if err != nil {
		t.Fatalf("CreatePaymentIntent failed: %v", err)
	}
	if pi.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if pi.ClientSecret == "" {
		t.Fatal("expected non-empty client secret")
	}
	if pi.Status != "requires_payment_method" {
		t.Fatalf("got %q, want requires_payment_method", pi.Status)
	}
	if pi.Amount != 1000 {
		t.Fatalf("got %d, want 1000", pi.Amount)
	}
	if pi.Currency != "usd" {
		t.Fatalf("got %q, want usd", pi.Currency)
	}
}

func TestCreatePaymentIntent_mock_zero_amount(t *testing.T) {
	c := &StripeClient{}
	_, err := c.CreatePaymentIntent(CreatePaymentIntentParams{Amount: 0})
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestCreatePaymentIntent_mock_negative_amount(t *testing.T) {
	c := &StripeClient{}
	_, err := c.CreatePaymentIntent(CreatePaymentIntentParams{Amount: -100})
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestConfirmPaymentIntent_mock(t *testing.T) {
	c := &StripeClient{}

	pi, err := c.ConfirmPaymentIntent("pi_mock_abc123", "pm_card_visa")
	if err != nil {
		t.Fatalf("ConfirmPaymentIntent failed: %v", err)
	}
	if pi.ID != "pi_mock_abc123" {
		t.Fatalf("got %q", pi.ID)
	}
	if pi.Status != "succeeded" {
		t.Fatalf("got %q, want succeeded", pi.Status)
	}
	if pi.Amount != 1000 {
		t.Fatalf("got %d", pi.Amount)
	}
	if pi.Currency != "usd" {
		t.Fatalf("got %q", pi.Currency)
	}
}

func TestConstructWebhookEvent_no_secret(t *testing.T) {
	c := &StripeClient{}

	eventPayload := map[string]string{"type": "payment_intent.succeeded"}
	payload, _ := json.Marshal(eventPayload)

	event, err := c.ConstructWebhookEvent(payload, "")
	if err != nil {
		t.Fatalf("ConstructWebhookEvent failed: %v", err)
	}
	if event.Type != "payment_intent.succeeded" {
		t.Fatalf("got %q", event.Type)
	}
}

func TestConstructWebhookEvent_invalid_payload(t *testing.T) {
	c := &StripeClient{}

	_, err := c.ConstructWebhookEvent([]byte("not json"), "")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestConstructWebhookEvent_with_secret_wrong_signature(t *testing.T) {
	c := &StripeClient{
		webhookSecret: "whsec_test_secret",
	}

	payload, _ := json.Marshal(map[string]string{"type": "test"})
	_, err := c.ConstructWebhookEvent(payload, "invalid-signature")
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestMockCreatePaymentIntent(t *testing.T) {
	c := &StripeClient{}
	params := CreatePaymentIntentParams{Amount: 500, Currency: "eur"}
	pi, err := c.mockCreatePaymentIntent(params)
	if err != nil {
		t.Fatalf("mockCreatePaymentIntent failed: %v", err)
	}
	if pi.Currency != "eur" {
		t.Fatalf("got %q", pi.Currency)
	}
	if pi.Amount != 500 {
		t.Fatalf("got %d", pi.Amount)
	}
}

func TestMockConfirmPaymentIntent(t *testing.T) {
	c := &StripeClient{}
	pi, err := c.mockConfirmPaymentIntent("pi_test", "pm_test")
	if err != nil {
		t.Fatalf("mockConfirmPaymentIntent failed: %v", err)
	}
	if pi.Status != "succeeded" {
		t.Fatalf("got %q", pi.Status)
	}
}

func TestErrWebhookSignatureInvalid(t *testing.T) {
	if !errors.Is(ErrWebhookSignatureInvalid, ErrWebhookSignatureInvalid) {
		t.Fatal("sentinel error mismatch")
	}
}

func TestErrPaymentIntentFailed(t *testing.T) {
	if !errors.Is(ErrPaymentIntentFailed, ErrPaymentIntentFailed) {
		t.Fatal("sentinel error mismatch")
	}
}
