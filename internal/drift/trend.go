package drift

import (
	"fmt"
	"sort"
	"time"
)

// TrendDirection indicates whether drift is increasing, decreasing, or stable.
type TrendDirection string

const (
	TrendIncreasing TrendDirection = "increasing"
	TrendDecreasing TrendDirection = "decreasing"
	TrendStable     TrendDirection = "stable"
)

// TrendPoint represents a single drift observation at a point in time.
type TrendPoint struct {
	Timestamp  time.Time
	Service    string
	DriftCount int
}

// TrendSummary summarises drift trend for a service over a window.
type TrendSummary struct {
	Service   string
	Points    []TrendPoint
	Direction TrendDirection
	Delta     int // latest minus earliest drift count
}

// String returns a human-readable description of the trend summary.
func (t TrendSummary) String() string {
	return fmt.Sprintf("service=%s direction=%s delta=%+d points=%d",
		t.Service, t.Direction, t.Delta, len(t.Points))
}

// TrendAnalyzer computes drift trends from a series of DriftResult values.
type TrendAnalyzer struct {
	points map[string][]TrendPoint
}

// NewTrendAnalyzer returns an initialised TrendAnalyzer.
func NewTrendAnalyzer() *TrendAnalyzer {
	return &TrendAnalyzer{points: make(map[string][]TrendPoint)}
}

// Record adds a DriftResult observation to the analyzer.
func (a *TrendAnalyzer) Record(r DriftResult, ts time.Time) {
	count := 0
	if r.HasDrift() {
		count = len(r.Entries)
	}
	a.points[r.Service] = append(a.points[r.Service], TrendPoint{
		Timestamp:  ts,
		Service:    r.Service,
		DriftCount: count,
	})
}

// Summarise returns a TrendSummary for the given service.
// Returns false if no data exists for the service.
func (a *TrendAnalyzer) Summarise(service string) (TrendSummary, bool) {
	pts, ok := a.points[service]
	if !ok || len(pts) == 0 {
		return TrendSummary{}, false
	}
	sorted := make([]TrendPoint, len(pts))
	copy(sorted, pts)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})
	first := sorted[0].DriftCount
	last := sorted[len(sorted)-1].DriftCount
	delta := last - first
	dir := TrendStable
	if delta > 0 {
		dir = TrendIncreasing
	} else if delta < 0 {
		dir = TrendDecreasing
	}
	return TrendSummary{
		Service:   service,
		Points:    sorted,
		Direction: dir,
		Delta:     delta,
	}, true
}

// Services returns all service names that have recorded points.
func (a *TrendAnalyzer) Services() []string {
	out := make([]string, 0, len(a.points))
	for svc := range a.points {
		out = append(out, svc)
	}
	sort.Strings(out)
	return out
}
