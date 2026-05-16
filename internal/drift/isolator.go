package drift

import "time"

// IsolationPolicy controls how long a service is isolated after repeated drift.
type IsolationPolicy struct {
	Threshold int           // number of consecutive drifts before isolation
	Duration  time.Duration // how long isolation lasts
}

// DefaultIsolationPolicy returns sensible defaults.
func DefaultIsolationPolicy() IsolationPolicy {
	return IsolationPolicy{
		Threshold: 3,
		Duration:  10 * time.Minute,
	}
}

type isolationState struct {
	consecutive int
	isolatedAt  time.Time
	isolated    bool
}

// Isolator tracks consecutive drift events per service and isolates
// services that exceed the configured threshold.
type Isolator struct {
	policy IsolationPolicy
	state  map[string]*isolationState
}

// NewIsolator creates an Isolator with the given policy. Zero-value policy
// fields fall back to DefaultIsolationPolicy.
func NewIsolator(policy IsolationPolicy) *Isolator {
	def := DefaultIsolationPolicy()
	if policy.Threshold <= 0 {
		policy.Threshold = def.Threshold
	}
	if policy.Duration <= 0 {
		policy.Duration = def.Duration
	}
	return &Isolator{
		policy: policy,
		state:  make(map[string]*isolationState),
	}
}

// Record registers a drift result for a service, updating consecutive counts
// and isolation status accordingly.
func (iso *Isolator) Record(result DriftResult) {
	st := iso.getOrCreate(result.Service)
	if result.HasDrift() {
		st.consecutive++
		if st.consecutive >= iso.policy.Threshold && !st.isolated {
			st.isolated = true
			st.isolatedAt = time.Now()
		}
	} else {
		st.consecutive = 0
		st.isolated = false
	}
}

// IsIsolated reports whether the named service is currently isolated.
func (iso *Isolator) IsIsolated(service string) bool {
	st, ok := iso.state[service]
	if !ok {
		return false
	}
	if st.isolated && time.Since(st.isolatedAt) >= iso.policy.Duration {
		st.isolated = false
		st.consecutive = 0
		return false
	}
	return st.isolated
}

// Reset clears isolation state for the named service.
func (iso *Isolator) Reset(service string) {
	delete(iso.state, service)
}

func (iso *Isolator) getOrCreate(service string) *isolationState {
	if st, ok := iso.state[service]; ok {
		return st
	}
	st := &isolationState{}
	iso.state[service] = st
	return st
}
