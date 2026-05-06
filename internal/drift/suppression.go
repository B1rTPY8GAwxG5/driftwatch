package drift

import (
	"time"
)

// SuppressionRule defines a rule that suppresses drift alerts for a service
// and optionally a specific drift kind, until a given expiry time.
type SuppressionRule struct {
	Service  string    `yaml:"service"`
	Kind     DriftKind `yaml:"kind,omitempty"`
	Expiry   time.Time `yaml:"expiry"`
	Reason   string    `yaml:"reason,omitempty"`
}

// IsExpired reports whether the suppression rule has passed its expiry time.
func (r SuppressionRule) IsExpired(now time.Time) bool {
	return now.After(r.Expiry)
}

// SuppressionStore holds a collection of suppression rules and evaluates
// whether a given drift entry should be suppressed.
type SuppressionStore struct {
	rules []SuppressionRule
	now   func() time.Time
}

// NewSuppressionStore creates a SuppressionStore with the provided rules.
func NewSuppressionStore(rules []SuppressionRule) *SuppressionStore {
	return &SuppressionStore{
		rules: rules,
		now:   time.Now,
	}
}

// Add appends a new suppression rule to the store.
func (s *SuppressionStore) Add(rule SuppressionRule) {
	s.rules = append(s.rules, rule)
}

// IsSuppressed returns true if the given service + kind combination is
// covered by a non-expired suppression rule.
func (s *SuppressionStore) IsSuppressed(service string, kind DriftKind) bool {
	now := s.now()
	for _, r := range s.rules {
		if r.IsExpired(now) {
			continue
		}
		if r.Service != service {
			continue
		}
		// An empty Kind means suppress all drift kinds for the service.
		if r.Kind == "" || r.Kind == kind {
			return true
		}
	}
	return false
}

// PruneExpired removes all rules that have already expired.
func (s *SuppressionStore) PruneExpired() {
	now := s.now()
	active := s.rules[:0]
	for _, r := range s.rules {
		if !r.IsExpired(now) {
			active = append(active, r)
		}
	}
	s.rules = active
}

// Rules returns a copy of the current suppression rules.
func (s *SuppressionStore) Rules() []SuppressionRule {
	out := make([]SuppressionRule, len(s.rules))
	copy(out, s.rules)
	return out
}
