package drift

import (
	"sync"
	"time"
)

// MuteRule silences drift results for a specific service and kind until a deadline.
type MuteRule struct {
	Service  string
	Kind     DriftKind
	Deadline time.Time
	Reason   string
}

// IsExpired reports whether the mute rule has passed its deadline.
func (r MuteRule) IsExpired(now time.Time) bool {
	return now.After(r.Deadline)
}

// Muter suppresses DriftResult entries based on active mute rules.
type Muter struct {
	mu    sync.RWMutex
	rules []MuteRule
}

// NewMuter returns an empty Muter.
func NewMuter() *Muter {
	return &Muter{}
}

// Add registers a new mute rule. Expired rules are pruned on each add.
func (m *Muter) Add(rule MuteRule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	active := m.rules[:0]
	for _, r := range m.rules {
		if !r.IsExpired(now) {
			active = append(active, r)
		}
	}
	m.rules = append(active, rule)
}

// IsMuted reports whether the given service and kind are currently muted.
func (m *Muter) IsMuted(service string, kind DriftKind) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	now := time.Now()
	for _, r := range m.rules {
		if r.IsExpired(now) {
			continue
		}
		if r.Service == service && (r.Kind == kind || r.Kind == "") {
			return true
		}
	}
	return false
}

// Apply removes muted entries from the result's Entries slice.
// If all entries are muted the result is returned with an empty Entries list.
func (m *Muter) Apply(result DriftResult) DriftResult {
	filtered := result.Entries[:0]
	for _, e := range result.Entries {
		if !m.IsMuted(result.Service, e.Kind) {
			filtered = append(filtered, e)
		}
	}
	result.Entries = filtered
	return result
}

// Len returns the number of currently stored (including potentially expired) rules.
func (m *Muter) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rules)
}
