package drift

import (
	"fmt"
	"sync"
	"time"
)

// EscalationLevel represents how severe an escalation is.
type EscalationLevel int

const (
	EscalationNone     EscalationLevel = iota
	EscalationWarning                  // drift persisted beyond warn threshold
	EscalationCritical                 // drift persisted beyond critical threshold
)

func (e EscalationLevel) String() string {
	switch e {
	case EscalationWarning:
		return "warning"
	case EscalationCritical:
		return "critical"
	default:
		return "none"
	}
}

// EscalationPolicy controls when drift escalates.
type EscalationPolicy struct {
	WarnAfter     time.Duration
	CriticalAfter time.Duration
}

// DefaultEscalationPolicy returns sensible defaults.
func DefaultEscalationPolicy() EscalationPolicy {
	return EscalationPolicy{
		WarnAfter:     5 * time.Minute,
		CriticalAfter: 30 * time.Minute,
	}
}

// escalationEntry tracks when drift was first seen for a service.
type escalationEntry struct {
	firstSeen time.Time
}

// Escalator tracks how long drift has persisted and assigns escalation levels.
type Escalator struct {
	mu     sync.Mutex
	policy EscalationPolicy
	state  map[string]escalationEntry
}

// NewEscalator creates an Escalator with the given policy.
func NewEscalator(policy EscalationPolicy) *Escalator {
	if policy.WarnAfter <= 0 {
		policy = DefaultEscalationPolicy()
	}
	return &Escalator{
		policy: policy,
		state:  make(map[string]escalationEntry),
	}
}

// Evaluate returns the EscalationLevel for the given DriftResult.
// Clean results clear any tracked state for the service.
func (e *Escalator) Evaluate(result DriftResult) EscalationLevel {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := result.Service
	if key == "" {
		return EscalationNone
	}

	if !result.HasDrift() {
		delete(e.state, key)
		return EscalationNone
	}

	entry, exists := e.state[key]
	if !exists {
		e.state[key] = escalationEntry{firstSeen: time.Now()}
		return EscalationNone
	}

	age := time.Since(entry.firstSeen)
	switch {
	case age >= e.policy.CriticalAfter:
		return EscalationCritical
	case age >= e.policy.WarnAfter:
		return EscalationWarning
	default:
		return EscalationNone
	}
}

// Reset clears tracked state for a specific service.
func (e *Escalator) Reset(service string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.state, service)
}

// Summary returns a human-readable description of the current escalation state.
func (e *Escalator) Summary() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return fmt.Sprintf("escalator: tracking %d service(s)", len(e.state))
}
