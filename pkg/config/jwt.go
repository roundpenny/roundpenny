package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func LoadJWTSecret() (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secretFile := os.Getenv("JWT_SECRET_FILE"); secretFile != "" {
		data, err := os.ReadFile(secretFile)
		if err == nil && len(data) > 0 {
			secret = strings.TrimSpace(string(data))
		}
	}

	isProd := os.Getenv("PRODUCTION") == "true"

	if secret == "" {
		if isProd {
			return "", fmt.Errorf("JWT_SECRET is required in production mode")
		}
		slog.Warn("JWT_SECRET not set, using insecure default")
		return "dev-secret-change-in-production", nil
	}

	if len(secret) < 32 {
		if isProd {
			return "", fmt.Errorf("JWT_SECRET must be at least 32 characters long (got %d)", len(secret))
		}
		slog.Warn("JWT_SECRET is shorter than 32 characters", "length", len(secret))
	}

	return secret, nil
}

func IsProduction() bool {
	return os.Getenv("PRODUCTION") == "true"
}

func ValidateCORSOrigins() []string {
	origins := os.Getenv("CORS_ORIGINS")
	if origins == "" || origins == "*" {
		if IsProduction() {
			slog.Warn("CORS_ORIGINS is set to wildcard in production mode, consider restricting")
		}
		return []string{"*"}
	}
	parts := strings.Split(origins, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
