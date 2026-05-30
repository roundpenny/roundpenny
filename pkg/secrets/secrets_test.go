// Copyright (c) 2026 RoundPenny. All rights reserved.

package secrets

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	vault "github.com/hashicorp/vault/api"
)

func TestEnvManager_Get(t *testing.T) {
	os.Setenv("TEST_SECRET_KEY", "super-secret-value")
	defer os.Unsetenv("TEST_SECRET_KEY")

	m := NewEnvManager()
	val, err := m.Get(context.Background(), "TEST_SECRET_KEY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "super-secret-value" {
		t.Fatalf("got %q, want %q", val, "super-secret-value")
	}
}

func TestEnvManager_Get_missing(t *testing.T) {
	m := NewEnvManager()
	_, err := m.Get(context.Background(), "NONEXISTENT_KEY")
	if err != ErrNotFound {
		t.Fatalf("got %v, want %v", err, ErrNotFound)
	}
}

func TestEnvManager_Get_empty_env_var(t *testing.T) {
	os.Setenv("EMPTY_SECRET", "")
	defer os.Unsetenv("EMPTY_SECRET")

	m := NewEnvManager()
	_, err := m.Get(context.Background(), "EMPTY_SECRET")
	if err != ErrNotFound {
		t.Fatalf("got %v, want %v", err, ErrNotFound)
	}
}

func TestEnvManager_Close(t *testing.T) {
	m := NewEnvManager()
	if err := m.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVaultManager_Get(t *testing.T) {
	srv := newVaultTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("X-Vault-Token", "test-token")
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"data": map[string]interface{}{
					"value": "db-password-123",
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	m := &VaultManager{
		client: newTestVaultClient(t, srv.URL),
		path:   "secret",
	}

	val, err := m.Get(context.Background(), "db_password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "db-password-123" {
		t.Fatalf("got %q, want %q", val, "db-password-123")
	}
}

func TestVaultManager_Get_not_found(t *testing.T) {
	srv := newVaultTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Vault-Token", "test-token")
		w.WriteHeader(http.StatusNotFound)
		resp := map[string]interface{}{
			"errors": []string{"no secret found"},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	m := &VaultManager{
		client: newTestVaultClient(t, srv.URL),
		path:   "secret",
	}

	_, err := m.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestManagerInterface(t *testing.T) {
	var _ Manager = (*EnvManager)(nil)
	var _ Manager = (*VaultManager)(nil)
}

func newVaultTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler(w, r)
	}))
}

func newTestVaultClient(t *testing.T, addr string) *vault.Client {
	t.Helper()
	config := vault.DefaultConfig()
	config.Address = addr
	client, err := vault.NewClient(config)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	client.SetToken("test-token")
	return client
}
