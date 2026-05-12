package drift

import "sort"

// Priority represents the urgency level assigned to a drift result.
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityMedium Priority = 2
	PriorityHigh   Priority = 3
	PriorityCritical Priority = 4
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// PrioritizedResult pairs a DriftResult with its computed priority.
type PrioritizedResult struct {
	Result   DriftResult
	Priority Priority
}

// Prioritizer assigns priorities to drift results based on configurable rules.
type Prioritizer struct {
	kindWeights map[DriftKind]Priority
}

// NewPrioritizer returns a Prioritizer with default kind weights.
func NewPrioritizer() *Prioritizer {
	return &Prioritizer{
		kindWeights: map[DriftKind]Priority{
			DriftKindImage:    PriorityHigh,
			DriftKindReplicas: PriorityMedium,
			DriftKindEnv:      PriorityLow,
		},
	}
}

// SetWeight overrides the priority weight for a specific DriftKind.
func (p *Prioritizer) SetWeight(kind DriftKind, priority Priority) {
	p.kindWeights[kind] = priority
}

// Prioritize computes the highest priority among all drift entries in the result.
func (p *Prioritizer) Prioritize(result DriftResult) PrioritizedResult {
	max := PriorityLow
	for _, entry := range result.Entries {
		if w, ok := p.kindWeights[entry.Kind]; ok && w > max {
			max = w
		}
	}
	if !result.HasDrift() {
		max = PriorityLow
	}
	return PrioritizedResult{Result: result, Priority: max}
}

// PrioritizeAll returns a slice of PrioritizedResults sorted by descending priority.
func (p *Prioritizer) PrioritizeAll(results []DriftResult) []PrioritizedResult {
	out := make([]PrioritizedResult, len(results))
	for i, r := range results {
		out[i] = p.Prioritize(r)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Priority > out[j].Priority
	})
	return out
}
