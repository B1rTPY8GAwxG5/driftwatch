package drift

import (
	"errors"
	"time"
)

// RetryPolicy defines how retries are attempted on transient failures.
type RetryPolicy struct {
	MaxAttempts int
	Delay       time.Duration
	Multiplier  float64
}

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		Delay:       200 * time.Millisecond,
		Multiplier:  2.0,
	}
}

// Validate returns an error if the policy is misconfigured.
func (r RetryPolicy) Validate() error {
	if r.MaxAttempts < 1 {
		return errors.New("retry: MaxAttempts must be at least 1")
	}
	if r.Delay < 0 {
		return errors.New("retry: Delay must be non-negative")
	}
	if r.Multiplier < 1.0 {
		return errors.New("retry: Multiplier must be >= 1.0")
	}
	return nil
}

// RetryDetector wraps a Detector and retries on error according to a RetryPolicy.
type RetryDetector struct {
	inner  Detector
	policy RetryPolicy
	sleep  func(time.Duration)
}

// NewRetryDetector creates a RetryDetector with the given policy.
// If policy is zero-valued, DefaultRetryPolicy is used.
func NewRetryDetector(inner Detector, policy RetryPolicy) (*RetryDetector, error) {
	if inner == nil {
		return nil, errors.New("retry: inner detector must not be nil")
	}
	if policy.MaxAttempts == 0 {
		policy = DefaultRetryPolicy()
	}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	return &RetryDetector{
		inner:  inner,
		policy: policy,
		sleep:  time.Sleep,
	}, nil
}

// Compare attempts the underlying detector up to MaxAttempts times,
// backing off between attempts. Returns the last error if all attempts fail.
func (r *RetryDetector) Compare(spec ServiceSpec, live ServiceSpec) (DriftResult, error) {
	delay := r.policy.Delay
	var lastErr error
	for attempt := 0; attempt < r.policy.MaxAttempts; attempt++ {
		result, err := r.inner.Compare(spec, live)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if attempt < r.policy.MaxAttempts-1 {
			r.sleep(delay)
			delay = time.Duration(float64(delay) * r.policy.Multiplier)
		}
	}
	return DriftResult{}, lastErr
}
