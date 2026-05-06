package drift

import (
	"fmt"
	"io"
	"sort"
)

// ScoreReport aggregates DriftScore values across multiple services.
type ScoreReport struct {
	Scores []DriftScore
}

// BuildScoreReport scores each result and returns a ScoreReport sorted by total descending.
func BuildScoreReport(results []DriftResult) ScoreReport {
	scores := make([]DriftScore, 0, len(results))
	for _, r := range results {
		scores = append(scores, ScoreResult(r))
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Total > scores[j].Total
	})
	return ScoreReport{Scores: scores}
}

// HasCritical reports whether any service scored at critical level.
func (sr ScoreReport) HasCritical() bool {
	for _, s := range sr.Scores {
		if s.Level == "critical" {
			return true
		}
	}
	return false
}

// WriteTo writes a human-readable score report to w.
func (sr ScoreReport) WriteTo(w io.Writer) error {
	if len(sr.Scores) == 0 {
		_, err := fmt.Fprintln(w, "score report: no results")
		return err
	}
	for _, s := range sr.Scores {
		line := fmt.Sprintf("service=%-20s score=%-4d level=%s\n", s.Service, s.Total, s.Level)
		if _, err := fmt.Fprint(w, line); err != nil {
			return err
		}
	}
	return nil
}
