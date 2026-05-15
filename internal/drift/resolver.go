package drift

import "fmt"

// ResolveStrategy determines how conflicting drift entries are resolved.
type ResolveStrategy string

const (
	ResolveStrategyLatest   ResolveStrategy = "latest"
	ResolveStrategyEarliest ResolveStrategy = "earliest"
	ResolveStrategySeverest ResolveStrategy = "severest"
)

// Resolver merges duplicate drift entries for the same service and kind,
// applying a configurable resolution strategy.
type Resolver struct {
	strategy ResolveStrategy
}

// NewResolver returns a Resolver with the given strategy.
// Falls back to ResolveStrategyLatest for unrecognised strategies.
func NewResolver(strategy ResolveStrategy) *Resolver {
	switch strategy {
	case ResolveStrategyLatest, ResolveStrategyEarliest, ResolveStrategySeverest:
	default:
		strategy = ResolveStrategyLatest
	}
	return &Resolver{strategy: strategy}
}

// Strategy returns the active resolution strategy.
func (r *Resolver) Strategy() ResolveStrategy { return r.strategy }

// Resolve deduplicates drift entries within result, keeping one entry per
// (kind, field) pair according to the configured strategy.
func (r *Resolver) Resolve(result DriftResult) DriftResult {
	type key struct {
		kind  DriftKind
		field string
	}

	seen := make(map[key]DriftEntry)

	for _, e := range result.Entries {
		k := key{kind: e.Kind, field: e.Field}
		existing, ok := seen[k]
		if !ok {
			seen[k] = e
			continue
		}
		switch r.strategy {
		case ResolveStrategyEarliest:
			if e.DetectedAt.Before(existing.DetectedAt) {
				seen[k] = e
			}
		case ResolveStrategySeverest:
			if severityRank(e.Kind) > severityRank(existing.Kind) {
				seen[k] = e
			}
		default: // latest
			if e.DetectedAt.After(existing.DetectedAt) {
				seen[k] = e
			}
		}
	}

	resolved := make([]DriftEntry, 0, len(seen))
	for _, e := range seen {
		resolved = append(resolved, e)
	}

	result.Entries = resolved
	return result
}

// ResolveAll applies Resolve to each result in the slice.
func (r *Resolver) ResolveAll(results []DriftResult) []DriftResult {
	out := make([]DriftResult, len(results))
	for i, res := range results {
		out[i] = r.Resolve(res)
	}
	return out
}

func severityRank(k DriftKind) int {
	switch k {
	case DriftKindImage:
		return 3
	case DriftKindReplicas:
		return 2
	case DriftKindEnv:
		return 1
	default:
		return 0
	}
}

// LoadResolverStrategy parses a strategy string and returns the typed constant.
func LoadResolverStrategy(s string) (ResolveStrategy, error) {
	switch ResolveStrategy(s) {
	case ResolveStrategyLatest, ResolveStrategyEarliest, ResolveStrategySeverest:
		return ResolveStrategy(s), nil
	}
	return "", fmt.Errorf("unknown resolve strategy: %q", s)
}
