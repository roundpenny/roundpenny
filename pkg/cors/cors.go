package cors

import (
	"net/http"
	"os"
	"strings"
)

func Middleware(next http.Handler) http.Handler {
	origins := os.Getenv("CORS_ORIGINS")
	if origins == "" {
		origins = "*"
	}
	allowedOrigins := strings.Split(origins, ",")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		matchedOrigin := ""
		for _, o := range allowedOrigins {
			o = strings.TrimSpace(o)
			if o == "*" || o == origin {
				matchedOrigin = o
				break
			}
		}

		if matchedOrigin == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if matchedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", matchedOrigin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID, Idempotency-Key")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, Idempotency-Key")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
