package drift

import (
	"sync"
	"time"
)

// CooldownPolicy defines the configuration for a cooldown period.
type CooldownPolicy struct {
	// Duration is the minimum time that must pass before the same service
	// can trigger another alert after a drift is detected.
	Duration time.Duration
}

// DefaultCooldownPolicy returns a CooldownPolicy with sensible defaults.
func DefaultCooldownPolicy() CooldownPolicy {
	return CooldownPolicy{
		Duration: 5 * time.Minute,
	}
}

// Cooldown tracks per-service cooldown state, suppressing repeated alerts
// within a configurable quiet period after a drift event.
type Cooldown struct {
	mu     sync.Mutex
	policy CooldownPolicy
	last   map[string]time.Time
	now    func() time.Time
}

// NewCooldown creates a new Cooldown with the given policy.
// If the policy duration is zero, DefaultCooldownPolicy is used.
func NewCooldown(policy CooldownPolicy) *Cooldown {
	if policy.Duration <= 0 {
		policy = DefaultCooldownPolicy()
	}
	return &Cooldown{
		policy: policy,
		last:   make(map[string]time.Time),
		now:    time.Now,
	}
}

// Allow returns true if the service is not currently in a cooldown period.
// If allowed, the cooldown timer for the service is reset.
func (c *Cooldown) Allow(service string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	if t, ok := c.last[service]; ok {
		if now.Sub(t) < c.policy.Duration {
			return false
		}
	}
	c.last[service] = now
	return true
}

// Reset clears the cooldown state for a specific service.
func (c *Cooldown) Reset(service string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.last, service)
}

// Flush clears all cooldown state.
func (c *Cooldown) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.last = make(map[string]time.Time)
}

// Active returns the number of services currently in a cooldown period.
func (c *Cooldown) Active() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	count := 0
	for _, t := range c.last {
		if now.Sub(t) < c.policy.Duration {
			count++
		}
	}
	return count
}
