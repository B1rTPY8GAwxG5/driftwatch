package drift

import (
	"fmt"
	"io"
	"sort"
	"time"
)

// ScorecardEntry holds a scored and labelled result for a single service.
type ScorecardEntry struct {
	Service   string
	Score     int
	Grade     string
	Drifted   bool
	Kinds     []string
	RecordedAt time.Time
}

// Scorecard aggregates scored entries across multiple services.
type Scorecard struct {
	entries []ScorecardEntry
}

// NewScorecard returns an empty Scorecard.
func NewScorecard() *Scorecard {
	return &Scorecard{}
}

// Add appends a ScorecardEntry to the scorecard.
func (s *Scorecard) Add(e ScorecardEntry) {
	if e.Service == "" {
		return
	}
	s.entries = append(s.entries, e)
}

// Entries returns a copy of all entries sorted by score descending.
func (s *Scorecard) Entries() []ScorecardEntry {
	out := make([]ScorecardEntry, len(s.entries))
	copy(out, s.entries)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Score > out[j].Score
	})
	return out
}

// Grade maps a numeric score to a letter grade.
func Grade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 75:
		return "B"
	case score >= 60:
		return "C"
	case score >= 40:
		return "D"
	default:
		return "F"
	}
}

// BuildScorecardEntry constructs a ScorecardEntry from a DriftResult.
func BuildScorecardEntry(r DriftResult, score int) ScorecardEntry {
	kinds := make([]string, 0, len(r.Entries))
	seen := map[string]bool{}
	for _, e := range r.Entries {
		k := string(e.Kind)
		if !seen[k] {
			kinds = append(kinds, k)
			seen[k] = true
		}
	}
	return ScorecardEntry{
		Service:    r.Service,
		Score:      score,
		Grade:      Grade(score),
		Drifted:    r.HasDrift(),
		Kinds:      kinds,
		RecordedAt: time.Now(),
	}
}

// WriteScorecardSummary writes a human-readable scorecard to w.
func WriteScorecardSummary(w io.Writer, sc *Scorecard) {
	entries := sc.Entries()
	if len(entries) == 0 {
		fmt.Fprintln(w, "scorecard: no entries")
		return
	}
	fmt.Fprintf(w, "%-30s %6s %5s %s\n", "SERVICE", "SCORE", "GRADE", "DRIFTED")
	for _, e := range entries {
		drifted := "no"
		if e.Drifted {
			drifted = "yes"
		}
		fmt.Fprintf(w, "%-30s %6d %5s %s\n", e.Service, e.Score, e.Grade, drifted)
	}
}
