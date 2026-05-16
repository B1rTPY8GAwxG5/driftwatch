package drift

import "fmt"

// SeverityRoute maps a severity level to a named destination.
type SeverityRoute struct {
	Severity string
	Dest     string
}

// SeverityRouter routes DriftResults to named destinations based on their
// computed severity level.
type SeverityRouter struct {
	routes   []SeverityRoute
	fallback string
}

// NewSeverityRouter creates a SeverityRouter with an optional fallback
// destination used when no route matches.
func NewSeverityRouter(fallback string) *SeverityRouter {
	return &SeverityRouter{fallback: fallback}
}

// AddRoute registers a severity-to-destination mapping.
// Duplicate severities are allowed; the first match wins.
func (r *SeverityRouter) AddRoute(severity, dest string) {
	if severity == "" || dest == "" {
		return
	}
	r.routes = append(r.routes, SeverityRoute{Severity: severity, Dest: dest})
}

// Route returns the destination for the given DriftResult.
// It uses the highest-scoring entry's severity to select a route.
// If no route matches, the fallback destination is returned.
func (r *SeverityRouter) Route(result DriftResult) string {
	if !result.HasDrift() {
		return r.fallback
	}
	severity := r.dominantSeverity(result)
	for _, route := range r.routes {
		if route.Severity == severity {
			return route.Dest
		}
	}
	return r.fallback
}

// Len returns the number of registered routes.
func (r *SeverityRouter) Len() int { return len(r.routes) }

// String returns a human-readable summary of the router configuration.
func (r *SeverityRouter) String() string {
	return fmt.Sprintf("SeverityRouter{routes:%d fallback:%q}", len(r.routes), r.fallback)
}

func (r *SeverityRouter) dominantSeverity(result DriftResult) string {
	order := map[string]int{"critical": 3, "high": 2, "medium": 1, "low": 0}
	best := -1
	sev := "low"
	for _, e := range result.Entries {
		s := string(e.Kind)
		if v, ok := order[s]; ok && v > best {
			best = v
			sev = s
		}
	}
	// Fall back to score-based severity when kinds don't map directly.
	scored := ScoreResult(result)
	if scored.Score >= 80 {
		return "critical"
	}
	if scored.Score >= 50 {
		return "high"
	}
	if scored.Score >= 20 {
		return "medium"
	}
	_ = sev
	return "low"
}
