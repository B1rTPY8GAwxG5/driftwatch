package drift

import (
	"fmt"
	"sync"
)

// DriftKey uniquely identifies a drift entry by service and kind.
type DriftKey struct {
	Service string
	Kind    DriftKind
	Field   string
}

// Deduplicator suppresses duplicate drift entries seen within a window.
type Deduplicator struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

// NewDeduplicator returns a new Deduplicator with an empty seen set.
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		seen: make(map[string]struct{}),
	}
}

// keyFor builds a stable string key from a DriftKey.
func keyFor(k DriftKey) string {
	return fmt.Sprintf("%s|%s|%s", k.Service, k.Kind, k.Field)
}

// IsDuplicate returns true if this DriftKey has already been seen.
// If it has not been seen it is recorded and false is returned.
func (d *Deduplicator) IsDuplicate(k DriftKey) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	key := keyFor(k)
	if _, exists := d.seen[key]; exists {
		return true
	}
	d.seen[key] = struct{}{}
	return false
}

// Reset clears all previously seen keys.
func (d *Deduplicator) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen = make(map[string]struct{})
}

// Deduplicate filters a DriftResult, removing entries already seen.
// Entries not yet seen are recorded and returned in a new DriftResult.
func (d *Deduplicator) Deduplicate(result DriftResult) DriftResult {
	var filtered []DriftEntry
	for _, entry := range result.Entries {
		k := DriftKey{
			Service: result.Service,
			Kind:    entry.Kind,
			Field:   entry.Field,
		}
		if !d.IsDuplicate(k) {
			filtered = append(filtered, entry)
		}
	}
	return DriftResult{
		Service: result.Service,
		Entries: filtered,
	}
}
