package monitoring

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsMiddleware_records_request_total(t *testing.T) {
	metrics := New("test-service")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	wrapped := MetricsMiddleware(metrics, handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got %d, want %d", rec.Code, http.StatusOK)
	}

	families, err := metrics.registry.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	var found bool
	for _, f := range families {
		if f.GetName() == "test_service_http_requests_total" {
			found = true
			if len(f.GetMetric()) == 0 {
				t.Fatal("expected at least one metric")
			}
			m := f.GetMetric()[0]
			if m.GetCounter().GetValue() < 1 {
				t.Fatalf("expected counter >= 1, got %f", m.GetCounter().GetValue())
			}
		}
	}
	if !found {
		t.Fatal("http_requests_total metric not found")
	}
}

func TestMetricsMiddleware_multiple_requests(t *testing.T) {
	metrics := New("test-service")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := MetricsMiddleware(metrics, inner)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/create", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
	}

	families, _ := metrics.registry.Gather()
	for _, f := range families {
		if f.GetName() == "test_service_http_requests_total" {
			total := 0.0
			for _, m := range f.GetMetric() {
				total += m.GetCounter().GetValue()
			}
			if total < 5 {
				t.Fatalf("expected >= 5 total requests, got %f", total)
			}
		}
	}
}

func TestMetricsMiddleware_records_duration(t *testing.T) {
	metrics := New("test-service")
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := MetricsMiddleware(metrics, inner)

	req := httptest.NewRequest("GET", "/slow", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	families, _ := metrics.registry.Gather()
	for _, f := range families {
		if f.GetName() == "test_service_http_request_duration_seconds" {
			if len(f.GetMetric()) == 0 {
				t.Fatal("expected duration metrics")
			}
			m := f.GetMetric()[0]
			if m.GetHistogram().GetSampleCount() < 1 {
				t.Fatalf("expected sample count >= 1, got %d", m.GetHistogram().GetSampleCount())
			}
		}
	}
}

func TestMetricsMiddleware_various_status_codes(t *testing.T) {
	metrics := New("test-service")

	codes := []int{http.StatusOK, http.StatusNotFound, http.StatusInternalServerError}
	for _, code := range codes {
		code := code
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
		})
		wrapped := MetricsMiddleware(metrics, inner)

		req := httptest.NewRequest("GET", "/resource", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
	}

	families, _ := metrics.registry.Gather()
	for _, f := range families {
		if f.GetName() == "test_service_http_requests_total" {
			if len(f.GetMetric()) == 0 {
				t.Fatal("expected metrics for various status codes")
			}
		}
	}
}
