package drift

import (
	"errors"
	"sync"
	"time"
)

// RateLimiter enforces a maximum number of drift checks per time window.
type RateLimiter struct {
	mu       sync.Mutex
	max      int
	window   time.Duration
	buckets  map[string][]time.Time
}

// NewRateLimiter creates a RateLimiter allowing at most max calls per window
// per service key. A zero or negative max defaults to 10.
func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	if max <= 0 {
		max = 10
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{
		max:    max,
		window: window,
		buckets: make(map[string][]time.Time),
	}
}

// Allow reports whether a check for the given service key is permitted under
// the current rate limit. It records the attempt if allowed.
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	times := r.buckets[key]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= r.max {
		r.buckets[key] = filtered
		return false
	}

	r.buckets[key] = append(filtered, now)
	return true
}

// Reset clears all recorded attempts for the given service key.
func (r *RateLimiter) Reset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.buckets, key)
}

// Count returns the number of recorded attempts within the current window for key.
func (r *RateLimiter) Count(key string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)
	count := 0
	for _, t := range r.buckets[key] {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}

// ErrRateLimited is returned when a service check is rejected by the limiter.
var ErrRateLimited = errors.New("drift: rate limit exceeded")
