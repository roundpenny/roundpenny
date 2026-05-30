// Copyright (c) 2026 RoundPenny. All rights reserved.

package db

import (
	"context"
	"errors"
	"testing"
)

func TestRetryWithBackoff_success(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	err := retryWithBackoff(ctx, func(ctx context.Context) error {
		callCount++
		return nil
	}, 3)

	if err != nil {
		t.Fatalf("retryWithBackoff failed: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 call, got %d", callCount)
	}
}

func TestRetryWithBackoff_eventual_success(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	err := retryWithBackoff(ctx, func(ctx context.Context) error {
		callCount++
		if callCount < 3 {
			return errors.New("transient error")
		}
		return nil
	}, 5)

	if err != nil {
		t.Fatalf("retryWithBackoff failed: %v", err)
	}
	if callCount != 3 {
		t.Fatalf("expected 3 calls, got %d", callCount)
	}
}

func TestRetryWithBackoff_all_fail(t *testing.T) {
	ctx := context.Background()
	errExpected := errors.New("persistent error")

	err := retryWithBackoff(ctx, func(ctx context.Context) error {
		return errExpected
	}, 3)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRetryWithBackoff_context_cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := retryWithBackoff(ctx, func(ctx context.Context) error {
		return errors.New("some error")
	}, 5)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}

func TestRetryWithBackoff_zero_retries(t *testing.T) {
	ctx := context.Background()
	errExpected := errors.New("fail")

	err := retryWithBackoff(ctx, func(ctx context.Context) error {
		return errExpected
	}, 1)

	if err == nil {
		t.Fatal("expected error")
	}
}

type mockPool struct {
	pingErr error
}

func (m *mockPool) Ping(ctx context.Context) error {
	return m.pingErr
}

func TestHealthCheck(t *testing.T) {
	t.Skip("Pool embeds pgxpool.Pool which cannot be mocked without pgx")
}
