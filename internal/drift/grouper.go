package drift

import "sort"

// GroupMode controls how results are grouped.
type GroupMode string

const (
	GroupByService GroupMode = "service"
	GroupByKind    GroupMode = "kind"
	GroupBySeverity GroupMode = "severity"
)

// ResultGroup holds a named collection of drift results.
type ResultGroup struct {
	Key     string
	Results []DriftResult
}

// Grouper partitions DriftResults into named buckets.
type Grouper struct {
	mode GroupMode
}

// NewGrouper returns a Grouper using the given mode.
// An unrecognised mode falls back to GroupByService.
func NewGrouper(mode GroupMode) *Grouper {
	switch mode {
	case GroupByService, GroupByKind, GroupBySeverity:
	default:
		mode = GroupByService
	}
	return &Grouper{mode: mode}
}

// Mode returns the active grouping mode.
func (g *Grouper) Mode() GroupMode { return g.mode }

// Group partitions results according to the grouper's mode.
// The returned slice is sorted by key for deterministic output.
func (g *Grouper) Group(results []DriftResult) []ResultGroup {
	buckets := make(map[string][]DriftResult)
	for _, r := range results {
		keys := g.keysFor(r)
		for _, k := range keys {
			buckets[k] = append(buckets[k], r)
		}
	}
	groups := make([]ResultGroup, 0, len(buckets))
	for k, rs := range buckets {
		groups = append(groups, ResultGroup{Key: k, Results: rs})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Key < groups[j].Key
	})
	return groups
}

func (g *Grouper) keysFor(r DriftResult) []string {
	switch g.mode {
	case GroupByKind:
		seen := map[string]bool{}
		for _, e := range r.Entries {
			k := string(e.Kind)
			if !seen[k] {
				seen[k] = true
			}
		}
		if len(seen) == 0 {
			return []string{"none"}
		}
		out := make([]string, 0, len(seen))
		for k := range seen {
			out = append(out, k)
		}
		return out
	case GroupBySeverity:
		_, level := ScoreResult(r)
		return []string{string(level)}
	default:
		if r.Service == "" {
			return []string{"unknown"}
		}
		return []string{r.Service}
	}
}
