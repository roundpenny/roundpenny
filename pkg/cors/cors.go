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
		for _, o := range allowedOrigins {
			o = strings.TrimSpace(o)
			if o == "*" || o == origin {
				w.Header().Set("Access-Control-Allow-Origin", o)
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
