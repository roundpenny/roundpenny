// Copyright (c) 2026 RoundPenny. All rights reserved.

package kyc

import (
	"testing"
)

func TestNewClient_mock(t *testing.T) {
	c := NewClient()
	if !c.mock {
		t.Fatal("expected mock mode when ONFIDO_API_TOKEN is not set")
	}
}

func TestCreateApplicant_mock(t *testing.T) {
	c := &Client{mock: true}

	a, err := c.CreateApplicant("John", "Doe", "john@test.com")
	if err != nil {
		t.Fatalf("CreateApplicant failed: %v", err)
	}
	if a.FirstName != "John" {
		t.Fatalf("got %q, want %q", a.FirstName, "John")
	}
	if a.LastName != "Doe" {
		t.Fatalf("got %q, want %q", a.LastName, "Doe")
	}
	if a.Email != "john@test.com" {
		t.Fatalf("got %q, want %q", a.Email, "john@test.com")
	}
	if a.ID == "" {
		t.Fatal("expected non-empty ID")
	}
}

func TestUploadDocument_mock(t *testing.T) {
	c := &Client{mock: true}

	docID, err := c.UploadDocument("app_123", "/path/to/file", "passport")
	if err != nil {
		t.Fatalf("UploadDocument failed: %v", err)
	}
	if docID == "" {
		t.Fatal("expected non-empty document ID")
	}
}

func TestCreateCheck_mock(t *testing.T) {
	c := &Client{mock: true}

	cr, err := c.CreateCheck("app_123", []string{"doc_123"})
	if err != nil {
		t.Fatalf("CreateCheck failed: %v", err)
	}
	if cr.ID == "" {
		t.Fatal("expected non-empty check ID")
	}
	if cr.Status != "complete" {
		t.Fatalf("got %q, want %q", cr.Status, "complete")
	}
	if cr.Result != "clear" {
		t.Fatalf("got %q, want %q", cr.Result, "clear")
	}
}

func TestMockCreateApplicant(t *testing.T) {
	c := &Client{mock: true}
	a, err := c.mockCreateApplicant("Alice", "Smith", "alice@test.com")
	if err != nil {
		t.Fatalf("mockCreateApplicant failed: %v", err)
	}
	if a.FirstName != "Alice" || a.LastName != "Smith" {
		t.Fatal("unexpected applicant data")
	}
}

func TestMockUploadDocument(t *testing.T) {
	c := &Client{mock: true}
	docID, err := c.mockUploadDocument("app_123")
	if err != nil {
		t.Fatalf("mockUploadDocument failed: %v", err)
	}
	if docID == "" {
		t.Fatal("expected non-empty document ID")
	}
}

func TestMockCreateCheck(t *testing.T) {
	c := &Client{mock: true}
	cr, err := c.mockCreateCheck("app_123")
	if err != nil {
		t.Fatalf("mockCreateCheck failed: %v", err)
	}
	if cr.Status != "complete" || cr.Result != "clear" {
		t.Fatal("unexpected check result")
	}
}
