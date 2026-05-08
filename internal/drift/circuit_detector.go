package drift

import "fmt"

// CircuitDetector wraps a Detector with a CircuitBreaker so that
// repeated downstream failures automatically pause drift checks until
// the circuit recovers.
type CircuitDetector struct {
	inner   Detector
	circuit *CircuitBreaker
}

// NewCircuitDetector returns a CircuitDetector that delegates to inner
// and is guarded by the provided CircuitBreaker.
func NewCircuitDetector(inner Detector, cb *CircuitBreaker) *CircuitDetector {
	if inner == nil {
		panic("circuit_detector: inner detector must not be nil")
	}
	if cb == nil {
		panic("circuit_detector: circuit breaker must not be nil")
	}
	return &CircuitDetector{inner: inner, circuit: cb}
}

// Compare delegates to the inner Detector if the circuit allows it.
// On success the circuit is notified; on error the failure is recorded.
func (cd *CircuitDetector) Compare(spec ServiceSpec, live ServiceSpec) (DriftResult, error) {
	if !cd.circuit.Allow() {
		return DriftResult{}, fmt.Errorf("%w: skipping drift check for %s", ErrCircuitOpen, spec.Name)
	}

	result, err := cd.inner.Compare(spec, live)
	if err != nil {
		cd.circuit.RecordFailure()
		return DriftResult{}, err
	}

	cd.circuit.RecordSuccess()
	return result, nil
}

// State exposes the underlying circuit state for observability.
func (cd *CircuitDetector) State() CircuitState {
	return cd.circuit.State()
}
