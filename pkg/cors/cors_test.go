package cors

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func resetCORSEnv() {
	os.Unsetenv("CORS_ORIGINS")
}

func TestMiddleware_passthrough(t *testing.T) {
	resetCORSEnv()
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

func TestMiddleware_options_preflight(t *testing.T) {
	resetCORSEnv()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called for OPTIONS")
	})
	wrapped := Middleware(handler)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestMiddleware_wildcard_origin(t *testing.T) {
	resetCORSEnv()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := Middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	h := rec.Header().Get("Access-Control-Allow-Origin")
	if h != "*" {
		t.Fatalf("got %q, want %q", h, "*")
	}
	if rec.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Fatal("credentials header should be empty for wildcard")
	}
}

func TestMiddleware_specific_origin_with_credentials(t *testing.T) {
	os.Setenv("CORS_ORIGINS", "http://example.com")
	defer resetCORSEnv()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := Middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	h := rec.Header().Get("Access-Control-Allow-Origin")
	if h != "http://example.com" {
		t.Fatalf("got %q, want %q", h, "http://example.com")
	}
	if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("expected credentials header")
	}
}

func TestMiddleware_headers_set(t *testing.T) {
	resetCORSEnv()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wrapped := Middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatal("expected Allow-Methods header")
	}
	if rec.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Fatal("expected Allow-Headers header")
	}
	if rec.Header().Get("Access-Control-Expose-Headers") == "" {
		t.Fatal("expected Expose-Headers header")
	}
	if rec.Header().Get("Access-Control-Max-Age") == "" {
		t.Fatal("expected Max-Age header")
	}
}

func TestMiddleware_missing_origin(t *testing.T) {
	resetCORSEnv()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := Middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("got %q, want %q", rec.Header().Get("Access-Control-Allow-Origin"), "*")
	}
}
