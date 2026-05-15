package drift

import (
	"fmt"
	"io"
	"sort"
	"time"
)

// ChangelogEntry records a single detected drift event for changelog purposes.
type ChangelogEntry struct {
	Service   string
	Kind      DriftKind
	Old       string
	New       string
	DetectedAt time.Time
}

// Changelog accumulates drift change entries over time.
type Changelog struct {
	entries []ChangelogEntry
}

// NewChangelog returns an empty Changelog.
func NewChangelog() *Changelog {
	return &Changelog{}
}

// Record appends drift entries from a DriftResult to the changelog.
func (c *Changelog) Record(result DriftResult) {
	if !result.HasDrift() {
		return
	}
	now := time.Now().UTC()
	for _, e := range result.Entries {
		c.entries = append(c.entries, ChangelogEntry{
			Service:    result.Service,
			Kind:       e.Kind,
			Old:        e.Declared,
			New:        e.Observed,
			DetectedAt: now,
		})
	}
}

// Entries returns a copy of all recorded changelog entries sorted by detection time.
func (c *Changelog) Entries() []ChangelogEntry {
	copy := make([]ChangelogEntry, len(c.entries))
	for i, e := range c.entries {
		copy[i] = e
	}
	sort.Slice(copy, func(i, j int) bool {
		return copy[i].DetectedAt.Before(copy[j].DetectedAt)
	})
	return copy
}

// Len returns the number of recorded entries.
func (c *Changelog) Len() int {
	return len(c.entries)
}

// WriteTo writes a human-readable changelog to w.
func (c *Changelog) WriteTo(w io.Writer) error {
	entries := c.Entries()
	if len(entries) == 0 {
		_, err := fmt.Fprintln(w, "changelog: no drift recorded")
		return err
	}
	for _, e := range entries {
		line := fmt.Sprintf("[%s] service=%s kind=%s declared=%q observed=%q\n",
			e.DetectedAt.Format(time.RFC3339), e.Service, e.Kind, e.Old, e.New)
		if _, err := fmt.Fprint(w, line); err != nil {
			return err
		}
	}
	return nil
}
