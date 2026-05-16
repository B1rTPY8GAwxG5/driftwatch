package drift

import (
	"fmt"
	"sync"
	"time"
)

// QuotaPeriod defines the window over which a quota is enforced.
type QuotaPeriod string

const (
	QuotaPeriodMinute QuotaPeriod = "minute"
	QuotaPeriodHour   QuotaPeriod = "hour"
	QuotaPeriodDay    QuotaPeriod = "day"
)

// QuotaPolicy configures the maximum allowed drift events per period per service.
type QuotaPolicy struct {
	Limit  int         `yaml:"limit"`
	Period QuotaPeriod `yaml:"period"`
}

// Validate returns an error if the policy is misconfigured.
func (p QuotaPolicy) Validate() error {
	if p.Limit <= 0 {
		return fmt.Errorf("quota limit must be greater than zero")
	}
	switch p.Period {
	case QuotaPeriodMinute, QuotaPeriodHour, QuotaPeriodDay:
		return nil
	default:
		return fmt.Errorf("unknown quota period: %q", p.Period)
	}
}

func (p QuotaPolicy) duration() time.Duration {
	switch p.Period {
	case QuotaPeriodMinute:
		return time.Minute
	case QuotaPeriodHour:
		return time.Hour
	case QuotaPeriodDay:
		return 24 * time.Hour
	default:
		return time.Hour
	}
}

type quotaBucket struct {
	count    int
	resetAt  time.Time
}

// QuotaEnforcer tracks and enforces per-service drift event quotas.
type QuotaEnforcer struct {
	mu     sync.Mutex
	policy QuotaPolicy
	buckets map[string]*quotaBucket
}

// NewQuotaEnforcer creates a new QuotaEnforcer with the given policy.
func NewQuotaEnforcer(policy QuotaPolicy) (*QuotaEnforcer, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	return &QuotaEnforcer{
		policy:  policy,
		buckets: make(map[string]*quotaBucket),
	}, nil
}

// Allow reports whether the service is within its quota and records the event.
func (q *QuotaEnforcer) Allow(service string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	b, ok := q.buckets[service]
	if !ok || now.After(b.resetAt) {
		q.buckets[service] = &quotaBucket{count: 1, resetAt: now.Add(q.policy.duration())}
		return true
	}
	if b.count >= q.policy.Limit {
		return false
	}
	b.count++
	return true
}

// Remaining returns how many events the service may still emit in the current window.
func (q *QuotaEnforcer) Remaining(service string) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	b, ok := q.buckets[service]
	if !ok || now.After(b.resetAt) {
		return q.policy.Limit
	}
	rem := q.policy.Limit - b.count
	if rem < 0 {
		return 0
	}
	return rem
}

// Reset clears the quota state for a specific service.
func (q *QuotaEnforcer) Reset(service string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.buckets, service)
}
