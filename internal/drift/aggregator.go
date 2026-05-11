package drift

import (
	"fmt"
	"sort"
	"strings"
)

// AggregatorMode controls how multiple DriftResults are combined.
type AggregatorMode string

const (
	AggregatorModeAny      AggregatorMode = "any"
	AggregatorModeAll      AggregatorMode = "all"
	AggregatorModeSummary  AggregatorMode = "summary"
)

// AggregatedResult holds the combined outcome of multiple DriftResults.
type AggregatedResult struct {
	Mode     AggregatorMode
	Results  []DriftResult
	Drifted  bool
	Services []string
}

// HasDrift returns true when the aggregated result contains drift.
func (a AggregatedResult) HasDrift() bool { return a.Drifted }

// Summary returns a human-readable description of the aggregated result.
func (a AggregatedResult) Summary() string {
	if !a.Drifted {
		return fmt.Sprintf("[%s] all %d service(s) clean", a.Mode, len(a.Results))
	}
	return fmt.Sprintf("[%s] drift detected in: %s", a.Mode, strings.Join(a.Services, ", "))
}

// Aggregator combines multiple DriftResults according to a mode.
type Aggregator struct {
	mode AggregatorMode
}

// NewAggregator returns an Aggregator using the given mode.
// Defaults to AggregatorModeAny when mode is empty.
func NewAggregator(mode AggregatorMode) *Aggregator {
	if mode == "" {
		mode = AggregatorModeAny
	}
	return &Aggregator{mode: mode}
}

// Aggregate combines results according to the configured mode.
func (a *Aggregator) Aggregate(results []DriftResult) AggregatedResult {
	out := AggregatedResult{Mode: a.mode, Results: results}
	var driftedServices []string
	for _, r := range results {
		if r.HasDrift() {
			driftedServices = append(driftedServices, r.Service)
		}
	}
	sort.Strings(driftedServices)
	out.Services = driftedServices
	switch a.mode {
	case AggregatorModeAll:
		out.Drifted = len(driftedServices) == len(results) && len(results) > 0
	default: // any, summary
		out.Drifted = len(driftedServices) > 0
	}
	return out
}
