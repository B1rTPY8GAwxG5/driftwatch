package drift

import (
	"errors"
	"testing"
	"time"
)

func TestDefaultRetryPolicy_Values(t *testing.T) {
	p := DefaultRetryPolicy()
	if p.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", p.MaxAttempts)
	}
	if p.Delay != 200*time.Millisecond {
		t.Errorf("unexpected Delay: %v", p.Delay)
	}
	if p.Multiplier != 2.0 {
		t.Errorf("expected Multiplier=2.0, got %f", p.Multiplier)
	}
}

func TestRetryPolicy_Validate_Valid(t *testing.T) {
	if err := DefaultRetryPolicy().Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestRetryPolicy_Validate_ZeroAttempts(t *testing.T) {
	p := RetryPolicy{MaxAttempts: 0, Delay: 0, Multiplier: 1.0}
	if err := p.Validate(); err == nil {
		t.Error("expected error for MaxAttempts=0")
	}
}

func TestRetryPolicy_Validate_NegativeDelay(t *testing.T) {
	p := RetryPolicy{MaxAttempts: 1, Delay: -1, Multiplier: 1.0}
	if err := p.Validate(); err == nil {
		t.Error("expected error for negative Delay")
	}
}

func TestRetryPolicy_Validate_MultiplierBelowOne(t *testing.T) {
	p := RetryPolicy{MaxAttempts: 1, Delay: 0, Multiplier: 0.5}
	if err := p.Validate(); err == nil {
		t.Error("expected error for Multiplier < 1.0")
	}
}

func TestNewRetryDetector_NilInner(t *testing.T) {
	_, err := NewRetryDetector(nil, DefaultRetryPolicy())
	if err == nil {
		t.Error("expected error for nil inner detector")
	}
}

func TestNewRetryDetector_InvalidPolicy(t *testing.T) {
	d, _ := NewDetector()
	_, err := NewRetryDetector(d, RetryPolicy{MaxAttempts: 0, Multiplier: 1.0})
	if err == nil {
		t.Error("expected error for invalid policy")
	}
}

func TestRetryDetector_SucceedsFirstAttempt(t *testing.T) {
	inner, _ := NewDetector()
	rd, err := NewRetryDetector(inner, DefaultRetryPolicy())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rd.sleep = func(time.Duration) {}
	spec := ServiceSpec{Name: "svc", Image: "img:1"}
	_, err = rd.Compare(spec, spec)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRetryDetector_RetriesOnError(t *testing.T) {
	calls := 0
	failErr := errors.New("transient")
	mock := &mockDetector{fn: func(a, b ServiceSpec) (DriftResult, error) {
		calls++
		if calls < 3 {
			return DriftResult{}, failErr
		}
		return DriftResult{Service: a.Name}, nil
	}}
	rd, _ := NewRetryDetector(mock, RetryPolicy{MaxAttempts: 3, Delay: 0, Multiplier: 1.0})
	rd.sleep = func(time.Duration) {}
	spec := ServiceSpec{Name: "svc", Image: "img:1"}
	result, err := rd.Compare(spec, spec)
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if result.Service != "svc" {
		t.Errorf("unexpected service: %s", result.Service)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetryDetector_ExhaustsAttempts(t *testing.T) {
	failErr := errors.New("always fails")
	mock := &mockDetector{fn: func(a, b ServiceSpec) (DriftResult, error) {
		return DriftResult{}, failErr
	}}
	rd, _ := NewRetryDetector(mock, RetryPolicy{MaxAttempts: 2, Delay: 0, Multiplier: 1.0})
	rd.sleep = func(time.Duration) {}
	spec := ServiceSpec{Name: "svc", Image: "img:1"}
	_, err := rd.Compare(spec, spec)
	if err == nil {
		t.Error("expected error after exhausting attempts")
	}
}

type mockDetector struct {
	fn func(a, b ServiceSpec) (DriftResult, error)
}

func (m *mockDetector) Compare(a, b ServiceSpec) (DriftResult, error) {
	return m.fn(a, b)
}
