package drift

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// SummaryPeriod represents the time window for a drift summary.
type SummaryPeriod string

const (
	SummaryPeriodHourly  SummaryPeriod = "hourly"
	SummaryPeriodDaily   SummaryPeriod = "daily"
	SummaryPeriodWeekly  SummaryPeriod = "weekly"
)

// ServiceSummary holds aggregated drift statistics for a single service.
type ServiceSummary struct {
	Service      string
	TotalChecks  int
	DriftedChecks int
	Kinds        map[DriftKind]int
	LastDriftAt  time.Time
}

// DriftSummary is the top-level summary across all services.
type DriftSummary struct {
	Period       SummaryPeriod
	GeneratedAt  time.Time
	Services     []ServiceSummary
}

// DriftRate returns the overall drift rate as a percentage.
func (s *DriftSummary) DriftRate() float64 {
	var total, drifted int
	for _, svc := range s.Services {
		total += svc.TotalChecks
		drifted += svc.DriftedChecks
	}
	if total == 0 {
		return 0
	}
	return float64(drifted) / float64(total) * 100
}

// Summarizer builds a DriftSummary from a collection of DriftResults.
type Summarizer struct {
	period  SummaryPeriod
	buckets map[string]*ServiceSummary
}

// NewSummarizer creates a new Summarizer for the given period.
func NewSummarizer(period SummaryPeriod) *Summarizer {
	if period == "" {
		period = SummaryPeriodDaily
	}
	return &Summarizer{
		period:  period,
		buckets: make(map[string]*ServiceSummary),
	}
}

// Record ingests a DriftResult into the summarizer.
func (s *Summarizer) Record(r DriftResult) {
	svc := r.Service
	if svc == "" {
		return
	}
	b, ok := s.buckets[svc]
	if !ok {
		b = &ServiceSummary{
			Service: svc,
			Kinds:   make(map[DriftKind]int),
		}
		s.buckets[svc] = b
	}
	b.TotalChecks++
	if r.HasDrift() {
		b.DriftedChecks++
		b.LastDriftAt = time.Now()
		for _, e := range r.Entries {
			b.Kinds[e.Kind]++
		}
	}
}

// Build returns the current DriftSummary.
func (s *Summarizer) Build() DriftSummary {
	svcs := make([]ServiceSummary, 0, len(s.buckets))
	for _, b := range s.buckets {
		svcs = append(svcs, *b)
	}
	sort.Slice(svcs, func(i, j int) bool {
		return svcs[i].Service < svcs[j].Service
	})
	return DriftSummary{
		Period:      s.period,
		GeneratedAt: time.Now(),
		Services:    svcs,
	}
}

// WriteSummary writes a human-readable summary to w.
func WriteSummary(w io.Writer, ds DriftSummary) {
	fmt.Fprintf(w, "Drift Summary [%s] — generated %s\n", ds.Period, ds.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "Overall drift rate: %.1f%%\n", ds.DriftRate())
	fmt.Fprintln(w, strings.Repeat("-", 40))
	for _, svc := range ds.Services {
		fmt.Fprintf(w, "  %-30s checks=%d drifted=%d\n", svc.Service, svc.TotalChecks, svc.DriftedChecks)
		for k, c := range svc.Kinds {
			fmt.Fprintf(w, "    kind=%-20s count=%d\n", k, c)
		}
	}
}
