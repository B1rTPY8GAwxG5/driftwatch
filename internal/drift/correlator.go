package drift

import (
	"fmt"
	"sync"
	"time"
)

// CorrelationGroup holds a set of DriftResults that share a common cause or pattern.
type CorrelationGroup struct {
	ID       string
	Kind     DriftKind
	Services []string
	Results  []DriftResult
	Detected time.Time
}

// Summary returns a human-readable description of the group.
func (g CorrelationGroup) Summary() string {
	return fmt.Sprintf("group=%s kind=%s services=%d detected=%s",
		g.ID, g.Kind, len(g.Services), g.Detected.Format(time.RFC3339))
}

// Correlator groups DriftResults by shared DriftKind across multiple services.
type Correlator struct {
	mu     sync.Mutex
	groups map[DriftKind]*CorrelationGroup
	clock  func() time.Time
}

// NewCorrelator returns a Correlator ready for use.
func NewCorrelator() *Correlator {
	return &Correlator{
		groups: make(map[DriftKind]*CorrelationGroup),
		clock:  time.Now,
	}
}

// Ingest adds a DriftResult to any matching correlation groups.
// A result contributes to a group for each DriftKind present in its entries.
func (c *Correlator) Ingest(result DriftResult) {
	if !result.HasDrift() {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	seen := make(map[DriftKind]bool)
	for _, entry := range result.Entries {
		if seen[entry.Kind] {
			continue
		}
		seen[entry.Kind] = true

		g, ok := c.groups[entry.Kind]
		if !ok {
			g = &CorrelationGroup{
				ID:       fmt.Sprintf("corr-%s", entry.Kind),
				Kind:     entry.Kind,
				Detected: c.clock(),
			}
			c.groups[entry.Kind] = g
		}
		g.Services = appendUnique(g.Services, result.Service)
		g.Results = append(g.Results, result)
	}
}

// Groups returns a snapshot of all current correlation groups.
func (c *Correlator) Groups() []CorrelationGroup {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]CorrelationGroup, 0, len(c.groups))
	for _, g := range c.groups {
		out = append(out, *g)
	}
	return out
}

// Reset clears all correlation state.
func (c *Correlator) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.groups = make(map[DriftKind]*CorrelationGroup)
}

func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}
