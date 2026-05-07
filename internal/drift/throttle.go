package drift

import (
	"sync"
	"time"
)

// ThrottlePolicy controls how frequently alerts are emitted per service.
type ThrottlePolicy struct {
	MinInterval time.Duration
}

// Throttle suppresses repeated notifications for the same service within a
// minimum interval, preventing alert storms during sustained drift.
type Throttle struct {
	mu     sync.Mutex
	last   map[string]time.Time
	policy ThrottlePolicy
}

// NewThrottle returns a Throttle with the given policy.
// If MinInterval is zero, a default of 1 minute is used.
func NewThrottle(p ThrottlePolicy) *Throttle {
	if p.MinInterval <= 0 {
		p.MinInterval = time.Minute
	}
	return &Throttle{
		last:   make(map[string]time.Time),
		policy: p,
	}
}

// Allow returns true if enough time has passed since the last notification
// for the given service name, and records the current time if so.
func (t *Throttle) Allow(service string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if last, ok := t.last[service]; ok {
		if now.Sub(last) < t.policy.MinInterval {
			return false
		}
	}
	t.last[service] = now
	return true
}

// Reset clears the recorded timestamp for a specific service.
func (t *Throttle) Reset(service string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.last, service)
}

// Flush clears all recorded timestamps.
func (t *Throttle) Flush() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.last = make(map[string]time.Time)
}
