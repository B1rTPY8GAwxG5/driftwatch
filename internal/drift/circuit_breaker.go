package drift

import (
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // normal operation
	CircuitOpen                         // failing, requests blocked
	CircuitHalfOpen                     // probing for recovery
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker protects a downstream call by tracking consecutive
// failures and opening the circuit when a threshold is exceeded.
type CircuitBreaker struct {
	mu           sync.Mutex
	state        CircuitState
	failures     int
	threshold    int
	resetTimeout time.Duration
	openedAt     time.Time
}

// NewCircuitBreaker returns a CircuitBreaker with the given failure
// threshold and reset timeout.
func NewCircuitBreaker(threshold int, resetTimeout time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 3
	}
	if resetTimeout <= 0 {
		resetTimeout = 30 * time.Second
	}
	return &CircuitBreaker{
		state:        CircuitClosed,
		threshold:    threshold,
		resetTimeout: resetTimeout,
	}
}

// Allow reports whether a call should be attempted given the current
// circuit state.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitOpen:
		if time.Since(cb.openedAt) >= cb.resetTimeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	default:
		return true
	}
}

// RecordSuccess resets the failure counter and closes the circuit.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = CircuitClosed
}

// RecordFailure increments the failure counter and opens the circuit
// when the threshold is reached.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.failures >= cb.threshold {
		cb.state = CircuitOpen
		cb.openedAt = time.Now()
	}
}

// State returns the current CircuitState.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// ErrCircuitOpen is returned when a call is blocked by an open circuit.
var ErrCircuitOpen = fmt.Errorf("circuit breaker is open")
