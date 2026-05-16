package drift

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// ProfileEntry records timing information for a single drift check operation.
type ProfileEntry struct {
	Service   string
	Operation string
	Duration  time.Duration
	Timestamp time.Time
}

// DriftProfiler collects and reports timing profiles for drift operations.
type DriftProfiler struct {
	mu      sync.Mutex
	entries []ProfileEntry
	maxSize int
}

// NewDriftProfiler creates a new DriftProfiler with the given maximum entry capacity.
func NewDriftProfiler(maxSize int) *DriftProfiler {
	if maxSize <= 0 {
		maxSize = 256
	}
	return &DriftProfiler{maxSize: maxSize}
}

// Record adds a profile entry for a completed operation.
func (p *DriftProfiler) Record(service, operation string, d time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.entries) >= p.maxSize {
		p.entries = p.entries[1:]
	}
	p.entries = append(p.entries, ProfileEntry{
		Service:   service,
		Operation: operation,
		Duration:  d,
		Timestamp: time.Now().UTC(),
	})
}

// Entries returns a copy of all recorded profile entries.
func (p *DriftProfiler) Entries() []ProfileEntry {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]ProfileEntry, len(p.entries))
	copy(out, p.entries)
	return out
}

// TopN returns the n slowest entries across all services, sorted descending by duration.
func (p *DriftProfiler) TopN(n int) []ProfileEntry {
	entries := p.Entries()
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Duration > entries[j].Duration
	})
	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n]
}

// Summary returns a human-readable summary of profiling data.
func (p *DriftProfiler) Summary() string {
	entries := p.Entries()
	if len(entries) == 0 {
		return "profiler: no entries recorded"
	}
	var total time.Duration
	for _, e := range entries {
		total += e.Duration
	}
	avg := total / time.Duration(len(entries))
	top := p.TopN(1)
	return fmt.Sprintf("profiler: entries=%d avg=%s slowest=%s (%s/%s)",
		len(entries), avg, top[0].Duration, top[0].Service, top[0].Operation)
}
