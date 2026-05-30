package security

import (
	"net/http"
	"strings"
)

type config struct {
	apiPrefixes []string
}

type Option func(*config)

func WithAPIPrefixes(prefixes ...string) Option {
	return func(c *config) {
		c.apiPrefixes = prefixes
	}
}

func Middleware(next http.Handler, opts ...Option) http.Handler {
	cfg := &config{
		apiPrefixes: []string{"/v1/", "/api/"},
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")

		for _, prefix := range cfg.apiPrefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				w.Header().Set("Cache-Control", "no-store")
				break
			}
		}

		next.ServeHTTP(w, r)
	})
}
