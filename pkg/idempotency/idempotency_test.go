package idempotency

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMockClient() *Client {
	return &Client{mock: true, data: make(map[string]*StoredResponse)}
}

func TestTryLock_success(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()

	locked, err := c.TryLock(ctx, "key1")
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if !locked {
		t.Fatal("expected lock to succeed")
	}
}

func TestTryLock_conflict(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()

	c.TryLock(ctx, "key1")
	locked, err := c.TryLock(ctx, "key1")
	if err != nil {
		t.Fatalf("TryLock failed: %v", err)
	}
	if locked {
		t.Fatal("expected lock to fail on conflict")
	}
}

func TestGetSet(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()

	sr := &StoredResponse{StatusCode: 200, Headers: map[string]string{"X-Custom": "val"}, Body: []byte("ok")}
	err := c.Set(ctx, "key1", sr)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := c.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.StatusCode != 200 || string(got.Body) != "ok" || got.Headers["X-Custom"] != "val" {
		t.Fatal("retrieved data mismatch")
	}
}

func TestGet_missing(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()

	_, err := c.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestDelete(t *testing.T) {
	c := newMockClient()
	ctx := context.Background()

	c.Set(ctx, "key1", &StoredResponse{StatusCode: 200})
	err := c.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = c.Get(ctx, "key1")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	if len(key) != 64 {
		t.Fatalf("got length %d, want 64", len(key))
	}
}

func TestMiddleware_passthrough_get(t *testing.T) {
	c := newMockClient()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("get"))
	})
	wrapped := Middleware(c, inner)

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || rec.Body.String() != "get" {
		t.Fatal("GET should pass through")
	}
}

func TestMiddleware_passthrough_no_header(t *testing.T) {
	c := newMockClient()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	})
	wrapped := Middleware(c, inner)

	req := httptest.NewRequest(http.MethodPost, "/resource", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated || rec.Body.String() != "created" {
		t.Fatal("POST without header should pass through")
	}
}

func TestMiddleware_idempotent_key_happy_path(t *testing.T) {
	c := newMockClient()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("done"))
	})
	wrapped := Middleware(c, inner)

	req := httptest.NewRequest(http.MethodPost, "/resource", nil)
	req.Header.Set(HeaderKey, "test-key-123")
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || rec.Body.String() != "done" {
		t.Fatalf("got %d %q, want 200 done", rec.Code, rec.Body.String())
	}
}

func TestMiddleware_replay_same_response(t *testing.T) {
	c := newMockClient()
	callCount := 0
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("done"))
	})
	wrapped := Middleware(c, inner)

	req1 := httptest.NewRequest(http.MethodPost, "/resource", nil)
	req1.Header.Set(HeaderKey, "replay-key")
	rec1 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/resource", nil)
	req2.Header.Set(HeaderKey, "replay-key")
	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req2)

	if callCount != 1 {
		t.Fatalf("expected 1 call, got %d", callCount)
	}
	if rec2.Body.String() != "done" {
		t.Fatalf("got %q, want %q", rec2.Body.String(), "done")
	}
}

func TestMiddleware_conflict_in_progress(t *testing.T) {
	c := newMockClient()
	c.data["locked-key"] = nil

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wrapped := Middleware(c, inner)

	req := httptest.NewRequest(http.MethodPost, "/resource", nil)
	req.Header.Set(HeaderKey, "locked-key")
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusConflict)
	}
}
