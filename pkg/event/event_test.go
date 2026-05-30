// Copyright (c) 2026 RoundPenny. All rights reserved.

package event

import (
	"testing"
	"time"
)

func TestEvent_creation(t *testing.T) {
	e := Event{
		ID:        "evt_123",
		Type:      "test.event",
		Source:    "test-service",
		Subject:   "test-subject",
		Data:      map[string]string{"key": "value"},
		Timestamp: time.Now(),
	}
	if e.ID != "evt_123" {
		t.Fatalf("got %q", e.ID)
	}
	if e.Type != "test.event" {
		t.Fatalf("got %q", e.Type)
	}
}

func TestTransactionSettled_creation(t *testing.T) {
	ts := TransactionSettled{
		TransactionID: "tx_123",
		UserID:        "user_456",
		MerchantID:    "merchant_789",
		Amount:        42.50,
		Currency:      "USD",
		ExternalTxID:  "ext_abc",
	}
	if ts.Amount != 42.50 {
		t.Fatalf("got %f", ts.Amount)
	}
}

func TestRoundUpCalculated_creation(t *testing.T) {
	ru := RoundUpCalculated{
		TransactionID:  "tx_123",
		UserID:         "user_456",
		OriginalAmount: 12.75,
		RoundedAmount:  13.00,
		RoundUpAmount:  0.25,
		Currency:       "USD",
	}
	if ru.RoundUpAmount != 0.25 {
		t.Fatalf("got %f", ru.RoundUpAmount)
	}
}

func TestWalletCredited_creation(t *testing.T) {
	wc := WalletCredited{
		UserID:    "user_123",
		Amount:    10.00,
		Reference: "ref_abc",
		Balance:   100.00,
	}
	if wc.Balance != 100.00 {
		t.Fatalf("got %f", wc.Balance)
	}
}

func TestFeeCharged_creation(t *testing.T) {
	fc := FeeCharged{
		TransactionID: "tx_123",
		UserID:        "user_456",
		Amount:        0.50,
		FeeType:       "processing",
	}
	if fc.FeeType != "processing" {
		t.Fatalf("got %q", fc.FeeType)
	}
}

func TestInvestmentCreated_creation(t *testing.T) {
	ic := InvestmentCreated{
		UserID:    "user_123",
		Portfolio: "aggressive-growth",
		Amount:    500.00,
	}
	if ic.Amount != 500.00 {
		t.Fatalf("got %f", ic.Amount)
	}
}

func TestConstants(t *testing.T) {
	if TopicTransactionSettled != "tx.settled" {
		t.Fatalf("got %q", TopicTransactionSettled)
	}
	if TopicRoundUpCalculated != "roundup.calculated" {
		t.Fatalf("got %q", TopicRoundUpCalculated)
	}
	if TopicWalletCredited != "wallet.credited" {
		t.Fatalf("got %q", TopicWalletCredited)
	}
	if TopicFeeCharged != "fee.charged" {
		t.Fatalf("got %q", TopicFeeCharged)
	}
	if TopicInvestmentCreated != "investment.created" {
		t.Fatalf("got %q", TopicInvestmentCreated)
	}
}
