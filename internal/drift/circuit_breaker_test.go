package drift

import (
	"testing"
	"time"
)

func TestNewCircuitBreaker_DefaultsOnZero(t *testing.T) {
	cb := NewCircuitBreaker(0, 0)
	if cb.threshold != 3 {
		t.Errorf("expected threshold 3, got %d", cb.threshold)
	}
	if cb.resetTimeout != 30*time.Second {
		t.Errorf("expected resetTimeout 30s, got %v", cb.resetTimeout)
	}
}

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)
	if cb.State() != CircuitClosed {
		t.Errorf("expected closed, got %s", cb.State())
	}
}

func TestCircuitBreaker_Allow_ClosedState(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)
	if !cb.Allow() {
		t.Error("expected Allow() == true when closed")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Minute)
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != CircuitClosed {
		t.Error("expected still closed after 2 failures")
	}
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Errorf("expected open after threshold, got %s", cb.State())
	}
}

func TestCircuitBreaker_BlocksWhenOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, time.Minute)
	cb.RecordFailure()
	if cb.Allow() {
		t.Error("expected Allow() == false when open")
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond)
	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	if !cb.Allow() {
		t.Error("expected Allow() == true after reset timeout")
	}
	if cb.State() != CircuitHalfOpen {
		t.Errorf("expected half-open, got %s", cb.State())
	}
}

func TestCircuitBreaker_RecordSuccess_Closes(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond)
	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	cb.Allow() // transition to half-open
	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Errorf("expected closed after success, got %s", cb.State())
	}
	if cb.failures != 0 {
		t.Errorf("expected failures reset to 0, got %d", cb.failures)
	}
}

func TestCircuitState_String(t *testing.T) {
	cases := []struct {
		state CircuitState
		want  string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half-open"},
		{CircuitState(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.state.String(); got != tc.want {
			t.Errorf("state %d: expected %q, got %q", tc.state, tc.want, got)
		}
	}
}
