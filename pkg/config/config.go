// Copyright (c) 2026 RoundPenny. All rights reserved.

package config

import (
	"context"
	"os"
	"strconv"
	"strings"
)

// SecretManager defines the interface for retrieving secrets.
// Compatible with secrets.Manager from github.com/roundup-platform/pkg/secrets.
type SecretManager interface {
	Get(ctx context.Context, key string) (string, error)
	Close() error
}

type Config struct {
	JWTSecret       string
	TLSCertFile     string
	TLSKeyFile      string
	CORSOrigins     []string
	DBMaxConns      int32
	DBMinConns      int32
	DBMaxLifetime   string
	DBMaxIdleTime   string
	KafkaBrokers    string
	KafkaTLSEnabled bool
	KafkaTLSCert    string
	KafkaTLSKey     string
	KafkaTLSCA      string
}

type optionState struct {
	secrets SecretManager
}

type Option func(*optionState)

func WithSecrets(sm SecretManager) Option {
	return func(s *optionState) {
		s.secrets = sm
	}
}

func Load(opts ...Option) *Config {
	state := new(optionState)
	for _, opt := range opts {
		opt(state)
	}

	cfg := &Config{
		JWTSecret:       getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		TLSCertFile:     getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:      getEnv("TLS_KEY_FILE", ""),
		CORSOrigins:     getEnvSlice("CORS_ORIGINS", []string{"*"}),
		DBMaxConns:      int32(getEnvInt("DB_MAX_CONNS", 25)),
		DBMinConns:      int32(getEnvInt("DB_MIN_CONNS", 5)),
		DBMaxLifetime:   getEnv("DB_MAX_LIFETIME", "30m"),
		DBMaxIdleTime:   getEnv("DB_MAX_IDLE_TIME", "5m"),
		KafkaBrokers:    getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTLSEnabled: getEnv("KAFKA_TLS_ENABLED", "false") == "true",
		KafkaTLSCert:    getEnv("KAFKA_TLS_CERT", ""),
		KafkaTLSKey:     getEnv("KAFKA_TLS_KEY", ""),
		KafkaTLSCA:      getEnv("KAFKA_TLS_CA", ""),
	}

	if state.secrets != nil {
		ctx := context.Background()
		if v, err := state.secrets.Get(ctx, "JWT_SECRET"); err == nil {
			cfg.JWTSecret = v
		}
	}

	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getEnvSlice(key string, def []string) []string {
	if v := os.Getenv(key); v != "" {
		parts := strings.Split(v, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	return def
}
