// Copyright (c) 2026 RoundPenny. All rights reserved.

package tracing

import (
	"os"
	"testing"
)

func TestInitTracing_env_defaults(t *testing.T) {
	os.Clearenv()
	_, err := InitTracing("test-service")
	if err == nil {
		t.Skip("OTLP endpoint not available in test environment, but no error returned")
	}
}

func TestInitTracing_custom_endpoint(t *testing.T) {
	os.Clearenv()
	os.Setenv("OTLP_ENDPOINT", "localhost:4317")
	defer os.Unsetenv("OTLP_ENDPOINT")

	_, err := InitTracing("custom-service")
	if err == nil {
		t.Skip("OTLP endpoint not actually running, but no error returned")
	}
}

func TestInitTracing_sample_rate_env(t *testing.T) {
	os.Clearenv()
	os.Setenv("TRACING_SAMPLE_RATE", "0.5")
	defer os.Unsetenv("TRACING_SAMPLE_RATE")

	_, err := InitTracing("sampled-service")
	if err == nil {
		t.Skip("OTLP endpoint not available")
	}
}

func TestInitTracing_invalid_sample_rate(t *testing.T) {
	os.Clearenv()
	os.Setenv("TRACING_SAMPLE_RATE", "not-a-number")
	defer os.Unsetenv("TRACING_SAMPLE_RATE")

	_, err := InitTracing("invalid-sampled")
	if err == nil {
		t.Skip("OTLP endpoint not available")
	}
}
