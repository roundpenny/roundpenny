// Copyright (c) 2026 RoundPenny. All rights reserved.

package testhelper

import (
	"context"
	"os"
	"testing"
)

func SkipIfNoRedis(t *testing.T) {
	t.Helper()
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("REDIS_ADDR not set, skipping test")
	}
}

func Context() context.Context {
	return context.Background()
}
