package drift

import (
	"time"
)

// PrunePolicy controls how old drift results are pruned.
type PrunePolicy struct {
	// MaxAge is the maximum age of a result before it is pruned.
	MaxAge time.Duration
	// MaxEntries is the maximum number of results to retain per service.
	// Zero means unlimited.
	MaxEntries int
}

// DefaultPrunePolicy returns a sensible default pruning policy.
func DefaultPrunePolicy() PrunePolicy {
	return PrunePolicy{
		MaxAge:     72 * time.Hour,
		MaxEntries: 100,
	}
}

// Pruner removes stale or excess drift results from a collection.
type Pruner struct {
	policy PrunePolicy
	now    func() time.Time
}

// NewPruner creates a Pruner with the given policy.
func NewPruner(policy PrunePolicy) *Pruner {
	return &Pruner{
		policy: policy,
		now:    time.Now,
	}
}

// Prune filters out results that are older than MaxAge or exceed MaxEntries
// per service. Results are assumed to be ordered oldest-first.
func (p *Pruner) Prune(results []DriftResult) []DriftResult {
	cutoff := p.now().Add(-p.policy.MaxAge)

	// Group by service, keeping only those within age limit.
	byService := make(map[string][]DriftResult)
	for _, r := range results {
		if r.Timestamp.Before(cutoff) {
			continue
		}
		byService[r.Service] = append(byService[r.Service], r)
	}

	// Apply per-service entry cap (keep most recent).
	var pruned []DriftResult
	for _, entries := range byService {
		if p.policy.MaxEntries > 0 && len(entries) > p.policy.MaxEntries {
			entries = entries[len(entries)-p.policy.MaxEntries:]
		}
		pruned = append(pruned, entries...)
	}
	return pruned
}

// PruneAll applies Prune across a map of service→results, returning a
// flattened slice of surviving results.
func (p *Pruner) PruneAll(m map[string][]DriftResult) []DriftResult {
	var all []DriftResult
	for _, results := range m {
		all = append(all, results...)
	}
	return p.Prune(all)
}
