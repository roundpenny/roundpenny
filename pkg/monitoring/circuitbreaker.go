package monitoring

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half_open"
	case StateOpen:
		return "open"
	}
	return "unknown"
}

type CircuitBreaker struct {
	mu               sync.RWMutex
	state            State
	failureCount     int
	successCount     int
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	lastStateChange  time.Time
	name             string
	stateGauge       *prometheus.GaugeVec
}

var ErrCircuitOpen = errors.New("circuit breaker is open")

func NewCircuitBreaker(name string, failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
		lastStateChange:  time.Now(),
	}
}

func (cb *CircuitBreaker) SetStateGauge(g *prometheus.GaugeVec) {
	cb.stateGauge = g
}

func (cb *CircuitBreaker) setState(s State) {
	cb.state = s
	cb.lastStateChange = time.Now()
	if cb.stateGauge != nil {
		for _, st := range []State{StateClosed, StateHalfOpen, StateOpen} {
			var v float64
			if st == s {
				v = 1
			}
			cb.stateGauge.WithLabelValues(cb.name, st.String()).Set(v)
		}
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.allow() {
		return ErrCircuitOpen
	}

	err := fn()
	cb.record(err)
	return err
}

func (cb *CircuitBreaker) allow() bool {
	cb.mu.RLock()
	state := cb.state
	lastChange := cb.lastStateChange
	cb.mu.RUnlock()

	if state == StateOpen {
		if time.Since(lastChange) > cb.timeout {
			cb.mu.Lock()
			cb.setState(StateHalfOpen)
			slog.Info("circuit breaker half-open", "name", cb.name)
			cb.mu.Unlock()
			return true
		}
		return false
	}

	return true
}

func (cb *CircuitBreaker) record(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.successCount = 0
		slog.Warn("circuit breaker failure", "name", cb.name, "count", cb.failureCount, "threshold", cb.failureThreshold)

		if cb.state == StateHalfOpen || (cb.state == StateClosed && cb.failureCount >= cb.failureThreshold) {
			cb.setState(StateOpen)
			cb.failureCount = 0
			slog.Warn("circuit breaker opened", "name", cb.name)
		}
	} else {
		cb.successCount++
		cb.failureCount = 0

		if cb.state == StateHalfOpen && cb.successCount >= cb.successThreshold {
			cb.setState(StateClosed)
			cb.successCount = 0
			slog.Info("circuit breaker closed", "name", cb.name)
		}
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) IsOpen() bool {
	return cb.State() == StateOpen
}
