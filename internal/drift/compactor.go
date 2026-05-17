package drift

import (
	"sort"
	"time"
)

// CompactPolicy controls how results are compacted.
type CompactPolicy struct {
	// MaxAge is the oldest result to retain.
	MaxAge time.Duration
	// MaxResults is the maximum number of results to keep per service.
	MaxResults int
	// KeepDriftedOnly, when true, discards clean results.
	KeepDriftedOnly bool
}

// DefaultCompactPolicy returns sensible compaction defaults.
func DefaultCompactPolicy() CompactPolicy {
	return CompactPolicy{
		MaxAge:          24 * time.Hour,
		MaxResults:      50,
		KeepDriftedOnly: false,
	}
}

// Compactor reduces a slice of DriftResults according to a CompactPolicy.
type Compactor struct {
	policy CompactPolicy
}

// NewCompactor creates a Compactor with the given policy.
// Zero-value fields are replaced with defaults.
func NewCompactor(p CompactPolicy) *Compactor {
	def := DefaultCompactPolicy()
	if p.MaxAge <= 0 {
		p.MaxAge = def.MaxAge
	}
	if p.MaxResults <= 0 {
		p.MaxResults = def.MaxResults
	}
	return &Compactor{policy: p}
}

// Compact filters and trims results according to the policy.
// Results are returned sorted newest-first.
func (c *Compactor) Compact(results []DriftResult) []DriftResult {
	if len(results) == 0 {
		return nil
	}

	cutoff := time.Now().Add(-c.policy.MaxAge)

	// Filter by age and optionally by drift status.
	filtered := results[:0:0]
	for _, r := range results {
		if r.Timestamp.Before(cutoff) {
			continue
		}
		if c.policy.KeepDriftedOnly && !r.HasDrift() {
			continue
		}
		filtered = append(filtered, r)
	}

	// Sort newest first.
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Trim to MaxResults.
	if len(filtered) > c.policy.MaxResults {
		filtered = filtered[:c.policy.MaxResults]
	}

	return filtered
}

// Policy returns the active CompactPolicy.
func (c *Compactor) Policy() CompactPolicy {
	return c.policy
}
