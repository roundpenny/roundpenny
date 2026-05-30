// Copyright (c) 2026 RoundPenny. All rights reserved.

package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_sets_security_headers(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := Middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	headers := map[string]string{
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Permissions-Policy":        "geolocation=(), microphone=(), camera=()",
		"Content-Security-Policy":   "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'",
	}

	for h, want := range headers {
		if got := rec.Header().Get(h); got != want {
			t.Errorf("%s header: got %q, want %q", h, got, want)
		}
	}
}

func TestMiddleware_api_cache_control(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := Middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("API route: got %q, want %q", got, "no-store")
	}
}

func TestMiddleware_non_api_no_cache_control(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := Middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/static/app.js", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Errorf("Non-API route: got %q, want empty", got)
	}
}

func TestMiddleware_passthrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	wrapped := Middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("got %q, want %q", rec.Body.String(), "ok")
	}
}

func TestMiddleware_custom_api_prefixes(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := Middleware(handler, WithAPIPrefixes("/custom-api/"))

	req := httptest.NewRequest(http.MethodGet, "/v1/users", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Errorf("default prefix not set: got %q, want empty", got)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/custom-api/data", nil)
	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req2)

	if got := rec2.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("custom prefix: got %q, want %q", got, "no-store")
	}
}
