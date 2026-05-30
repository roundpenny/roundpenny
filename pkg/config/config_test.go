// Copyright (c) 2026 RoundPenny. All rights reserved.

package config

import (
	"os"
	"testing"
)

func TestLoad_defaults(t *testing.T) {
	os.Clearenv()
	cfg := Load()

	if cfg.JWTSecret != "dev-secret-change-in-production" {
		t.Fatalf("got %q, want default", cfg.JWTSecret)
	}
	if len(cfg.CORSOrigins) != 1 || cfg.CORSOrigins[0] != "*" {
		t.Fatalf("got %v, want [*]", cfg.CORSOrigins)
	}
	if cfg.DBMaxConns != 25 {
		t.Fatalf("got %d, want 25", cfg.DBMaxConns)
	}
	if cfg.DBMinConns != 5 {
		t.Fatalf("got %d, want 5", cfg.DBMinConns)
	}
	if cfg.DBMaxLifetime != "30m" {
		t.Fatalf("got %q, want 30m", cfg.DBMaxLifetime)
	}
	if cfg.DBMaxIdleTime != "5m" {
		t.Fatalf("got %q, want 5m", cfg.DBMaxIdleTime)
	}
	if cfg.KafkaBrokers != "localhost:9092" {
		t.Fatalf("got %q, want localhost:9092", cfg.KafkaBrokers)
	}
	if cfg.KafkaTLSEnabled {
		t.Fatal("expected TLS disabled by default")
	}
}

func TestLoad_with_env_vars(t *testing.T) {
	os.Clearenv()
	os.Setenv("JWT_SECRET", "custom-secret-here-which-is-long-enough!!")
	os.Setenv("CORS_ORIGINS", "https://app.example.com,https://admin.example.com")
	os.Setenv("DB_MAX_CONNS", "10")
	os.Setenv("DB_MIN_CONNS", "2")
	os.Setenv("KAFKA_BROKERS", "kafka:9092")
	os.Setenv("KAFKA_TLS_ENABLED", "true")
	os.Setenv("TLS_CERT_FILE", "/certs/tls.crt")
	os.Setenv("TLS_KEY_FILE", "/certs/tls.key")

	cfg := Load()

	if cfg.JWTSecret != "custom-secret-here-which-is-long-enough!!" {
		t.Fatalf("got %q", cfg.JWTSecret)
	}
	if len(cfg.CORSOrigins) != 2 {
		t.Fatalf("got %d origins", len(cfg.CORSOrigins))
	}
	if cfg.CORSOrigins[0] != "https://app.example.com" {
		t.Fatalf("got %q", cfg.CORSOrigins[0])
	}
	if cfg.CORSOrigins[1] != "https://admin.example.com" {
		t.Fatalf("got %q", cfg.CORSOrigins[1])
	}
	if cfg.DBMaxConns != 10 {
		t.Fatalf("got %d", cfg.DBMaxConns)
	}
	if cfg.DBMinConns != 2 {
		t.Fatalf("got %d", cfg.DBMinConns)
	}
	if cfg.KafkaBrokers != "kafka:9092" {
		t.Fatalf("got %q", cfg.KafkaBrokers)
	}
	if !cfg.KafkaTLSEnabled {
		t.Fatal("expected TLS enabled")
	}
	if cfg.TLSCertFile != "/certs/tls.crt" {
		t.Fatalf("got %q", cfg.TLSCertFile)
	}
	if cfg.TLSKeyFile != "/certs/tls.key" {
		t.Fatalf("got %q", cfg.TLSKeyFile)
	}
}

func TestLoad_invalid_int_falls_back(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_MAX_CONNS", "not-a-number")
	cfg := Load()
	if cfg.DBMaxConns != 25 {
		t.Fatalf("got %d, want 25", cfg.DBMaxConns)
	}
}

func TestLoad_empty_cors_defaults_to_wildcard(t *testing.T) {
	os.Clearenv()
	os.Setenv("CORS_ORIGINS", "")
	cfg := Load()
	if len(cfg.CORSOrigins) != 1 || cfg.CORSOrigins[0] != "*" {
		t.Fatal("expected wildcard default")
	}
}

func TestLoad_trimmed_cors_origins(t *testing.T) {
	os.Clearenv()
	os.Setenv("CORS_ORIGINS", "  https://a.com , https://b.com  ")
	cfg := Load()
	if cfg.CORSOrigins[0] != "https://a.com" || cfg.CORSOrigins[1] != "https://b.com" {
		t.Fatalf("got %v", cfg.CORSOrigins)
	}
}
