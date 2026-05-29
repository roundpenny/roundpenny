package audit

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const (
	ContextKeyRequestID contextKey = "request_id"
	ContextKeyClientIP  contextKey = "client_ip"
	ContextKeyUserAgent contextKey = "user_agent"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Middleware(logger *AuditLogger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set("X-Request-ID", requestID)
		}
		w.Header().Set("X-Request-ID", requestID)

		ctx := context.WithValue(r.Context(), ContextKeyRequestID, requestID)
		ctx = context.WithValue(ctx, ContextKeyClientIP, r.RemoteAddr)
		ctx = context.WithValue(ctx, ContextKeyUserAgent, r.UserAgent())
		r = r.WithContext(ctx)

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		userID := ""
		if uid, ok := ctx.Value("user_id").(string); ok {
			userID = uid
		}

		logger.Log(ctx, ActionLogin, "http_request", r.URL.Path, userID, map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": rw.statusCode,
			"duration_ms": duration.Milliseconds(),
			"duration":    duration.String(),
		})
	})
}
