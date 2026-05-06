package drift

import "fmt"

// RollupSeverity represents the aggregate severity of a rollup.
type RollupSeverity string

const (
	RollupClean    RollupSeverity = "clean"
	RollupLow      RollupSeverity = "low"
	RollupMedium   RollupSeverity = "medium"
	RollupHigh     RollupSeverity = "high"
	RollupCritical RollupSeverity = "critical"
)

// RollupEntry summarises drift for a single service.
type RollupEntry struct {
	Service  string
	Score    int
	Severity RollupSeverity
	Drifted  bool
	Kinds    []DriftKind
}

// Rollup aggregates drift results across multiple services.
type Rollup struct {
	Entries  []RollupEntry
	Total    int
	Drifted  int
	Severity RollupSeverity
}

// BuildRollup creates a Rollup from a slice of DriftResults.
func BuildRollup(results []DriftResult) Rollup {
	r := Rollup{Total: len(results)}
	maxScore := 0

	for _, res := range results {
		sr := ScoreResult(res)
		entry := RollupEntry{
			Service:  res.Service,
			Score:    sr.Score,
			Severity: rollupSeverity(sr.Score),
			Drifted:  res.HasDrift(),
		}
		for _, e := range res.Entries {
			entry.Kinds = append(entry.Kinds, e.Kind)
		}
		if res.HasDrift() {
			r.Drifted++
		}
		if sr.Score > maxScore {
			maxScore = sr.Score
		}
		r.Entries = append(r.Entries, entry)
	}

	r.Severity = rollupSeverity(maxScore)
	return r
}

// rollupSeverity maps a numeric score to a RollupSeverity.
func rollupSeverity(score int) RollupSeverity {
	switch {
	case score == 0:
		return RollupClean
	case score < 20:
		return RollupLow
	case score < 50:
		return RollupMedium
	case score < 80:
		return RollupHigh
	default:
		return RollupCritical
	}
}

// Summary returns a human-readable one-line summary of the rollup.
func (r Rollup) Summary() string {
	return fmt.Sprintf("services=%d drifted=%d severity=%s", r.Total, r.Drifted, r.Severity)
}
