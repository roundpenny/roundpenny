// Copyright (c) 2026 RoundPenny. All rights reserved.

package monitoring

import (
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 2, time.Second)
	if cb.State() != StateClosed {
		t.Fatalf("got %v, want closed", cb.State())
	}
	if cb.name != "test" {
		t.Fatalf("got %q, want %q", cb.name, "test")
	}
}

func TestExecute_success(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 2, time.Second)

	err := cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if cb.State() != StateClosed {
		t.Fatalf("got %v, want closed", cb.State())
	}
}

func TestExecute_failure_opens_circuit(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 1, time.Second)
	errFail := errors.New("service error")

	cb.Execute(func() error { return errFail })
	cb.Execute(func() error { return errFail })

	if cb.State() != StateOpen {
		t.Fatalf("got %v, want open", cb.State())
	}
}

func TestExecute_circuit_open_returns_error(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 1, time.Minute)

	cb.Execute(func() error { return errors.New("fail") })

	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("got %v, want ErrCircuitOpen", err)
	}
}

func TestExecute_half_open_retry_after_timeout(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 1, 50*time.Millisecond)

	cb.Execute(func() error { return errors.New("fail") })
	if cb.State() != StateOpen {
		t.Fatal("expected circuit to be open")
	}

	time.Sleep(60 * time.Millisecond)

	successCount := 0
	err := cb.Execute(func() error {
		successCount++
		return nil
	})
	if err != nil {
		t.Fatalf("Execute after timeout failed: %v", err)
	}
}

func TestExecute_half_open_to_closed(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 2, 50*time.Millisecond)

	cb.Execute(func() error { return errors.New("fail") })
	time.Sleep(60 * time.Millisecond)

	cb.Execute(func() error { return nil })
	if cb.State() != StateHalfOpen {
		t.Fatalf("got %v, want half_open", cb.State())
	}

	cb.Execute(func() error { return nil })
	if cb.State() != StateClosed {
		t.Fatalf("got %v, want closed", cb.State())
	}
}

func TestExecute_half_open_failure_opens_again(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 1, 50*time.Millisecond)

	cb.Execute(func() error { return errors.New("fail") })
	time.Sleep(60 * time.Millisecond)

	cb.Execute(func() error { return errors.New("fail again") })
	if cb.State() != StateOpen {
		t.Fatalf("got %v, want open", cb.State())
	}
}

func TestState_string(t *testing.T) {
	tests := []struct {
		s    State
		want string
	}{
		{StateClosed, "closed"},
		{StateHalfOpen, "half_open"},
		{StateOpen, "open"},
		{State(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.s.String(); got != tt.want {
			t.Fatalf("got %q, want %q", got, tt.want)
		}
	}
}

func TestIsOpen(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 1, time.Minute)
	if cb.IsOpen() {
		t.Fatal("expected closed initially")
	}

	cb.Execute(func() error { return errors.New("fail") })
	if !cb.IsOpen() {
		t.Fatal("expected open after failure")
	}
}

func TestSetStateGauge(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 1, time.Minute)
	reg := prometheus.NewRegistry()
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "test_gauge", Help: "test"},
		[]string{"name", "state"},
	)
	reg.MustRegister(gauge)
	cb.SetStateGauge(gauge)

	cb.Execute(func() error { return errors.New("fail") })

	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}
	var found bool
	for _, f := range families {
		if f.GetName() == "test_gauge" {
			found = true
		}
	}
	if !found {
		t.Fatal("gauge metric not found")
	}
}
