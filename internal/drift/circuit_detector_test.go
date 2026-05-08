package drift

import (
	"errors"
	"testing"
	"time"
)

// errDetector always returns an error from Compare.
type errDetector struct{ err error }

func (e *errDetector) Compare(_, _ ServiceSpec) (DriftResult, error) {
	return DriftResult{}, e.err
}

func TestNewCircuitDetector_NotNil(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)
	det := NewDetector()
	cd := NewCircuitDetector(det, cb)
	if cd == nil {
		t.Fatal("expected non-nil CircuitDetector")
	}
}

func TestCircuitDetector_Compare_DelegatesOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)
	det := NewDetector()
	cd := NewCircuitDetector(det, cb)

	spec := ServiceSpec{Name: "svc", Image: "nginx:1.25", Replicas: 2}
	result, err := cd.Compare(spec, spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasDrift() {
		t.Error("expected no drift")
	}
	if cd.State() != CircuitClosed {
		t.Errorf("expected closed circuit, got %s", cd.State())
	}
}

func TestCircuitDetector_RecordsFailure(t *testing.T) {
	cb := NewCircuitBreaker(2, time.Minute)
	inner := &errDetector{err: errors.New("downstream unavailable")}
	cd := NewCircuitDetector(inner, cb)

	spec := ServiceSpec{Name: "svc"}
	_, err := cd.Compare(spec, spec)
	if err == nil {
		t.Fatal("expected error from inner detector")
	}
	if cb.failures != 1 {
		t.Errorf("expected 1 failure recorded, got %d", cb.failures)
	}
}

func TestCircuitDetector_BlocksWhenOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, time.Minute)
	inner := &errDetector{err: errors.New("fail")}
	cd := NewCircuitDetector(inner, cb)

	spec := ServiceSpec{Name: "svc"}
	// trip the circuit
	_, _ = cd.Compare(spec, spec)

	if cd.State() != CircuitOpen {
		t.Fatalf("expected open circuit, got %s", cd.State())
	}

	_, err := cd.Compare(spec, spec)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitDetector_RecoverAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond)
	inner := &errDetector{err: errors.New("fail")}
	cd := NewCircuitDetector(inner, cb)

	spec := ServiceSpec{Name: "svc"}
	_, _ = cd.Compare(spec, spec) // open circuit

	time.Sleep(20 * time.Millisecond)

	// Now swap inner for a healthy detector
	cd.inner = NewDetector()
	_, err := cd.Compare(spec, spec)
	if err != nil {
		t.Fatalf("expected recovery, got error: %v", err)
	}
	if cd.State() != CircuitClosed {
		t.Errorf("expected closed after recovery, got %s", cd.State())
	}
}
