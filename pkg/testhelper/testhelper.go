package testhelper

import (
	"math"
	"testing"
)

func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func AssertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func AssertNotEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got == want {
		t.Fatalf("expected different value, got %v", want)
	}
}

func AlmostEqual(t *testing.T, a, b, eps float64) {
	t.Helper()
	if math.Abs(a-b) > eps {
		t.Fatalf("got %v, want %v (eps %v)", a, b, eps)
	}
}

func AssertTrue(t *testing.T, cond bool, msg string) {
	t.Helper()
	if !cond {
		t.Fatalf("expected true: %s", msg)
	}
}

func AssertFalse(t *testing.T, cond bool, msg string) {
	t.Helper()
	if cond {
		t.Fatalf("expected false: %s", msg)
	}
}
