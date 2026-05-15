package drift

import (
	"fmt"
	"sync"
	"time"
)

// VersionEntry records a single observed version of a service field.
type VersionEntry struct {
	Service   string
	Field     string
	Value     string
	ObservedAt time.Time
}

// VersionTracker tracks the history of field values across drift results,
// allowing callers to detect when a field has changed version.
type VersionTracker struct {
	mu      sync.Mutex
	history map[string][]VersionEntry
	maxAge  time.Duration
}

// NewVersionTracker returns a VersionTracker that retains entries up to maxAge.
// If maxAge is zero, a default of 24 hours is used.
func NewVersionTracker(maxAge time.Duration) *VersionTracker {
	if maxAge <= 0 {
		maxAge = 24 * time.Hour
	}
	return &VersionTracker{
		history: make(map[string][]VersionEntry),
		maxAge:  maxAge,
	}
}

func versionKey(service, field string) string {
	return fmt.Sprintf("%s::%s", service, field)
}

// Record stores the current value for a service field observed in result.
func (vt *VersionTracker) Record(result DriftResult) {
	vt.mu.Lock()
	defer vt.mu.Unlock()

	now := time.Now()
	for _, entry := range result.Entries {
		key := versionKey(result.Service, string(entry.Kind))
		vt.history[key] = append(vt.history[key], VersionEntry{
			Service:    result.Service,
			Field:      string(entry.Kind),
			Value:      entry.Got,
			ObservedAt: now,
		})
	}
	vt.evict(now)
}

// History returns all recorded entries for the given service and field.
func (vt *VersionTracker) History(service, field string) []VersionEntry {
	vt.mu.Lock()
	defer vt.mu.Unlock()

	key := versionKey(service, field)
	src := vt.history[key]
	out := make([]VersionEntry, len(src))
	copy(out, src)
	return out
}

// HasChanged reports whether the value for a service field differs from the
// most recently recorded value.
func (vt *VersionTracker) HasChanged(service, field, current string) bool {
	vt.mu.Lock()
	defer vt.mu.Unlock()

	key := versionKey(service, field)
	entries := vt.history[key]
	if len(entries) == 0 {
		return false
	}
	return entries[len(entries)-1].Value != current
}

// evict removes entries older than maxAge. Must be called with mu held.
func (vt *VersionTracker) evict(now time.Time) {
	cutoff := now.Add(-vt.maxAge)
	for key, entries := range vt.history {
		var kept []VersionEntry
		for _, e := range entries {
			if e.ObservedAt.After(cutoff) {
				kept = append(kept, e)
			}
		}
		if len(kept) == 0 {
			delete(vt.history, key)
		} else {
			vt.history[key] = kept
		}
	}
}
