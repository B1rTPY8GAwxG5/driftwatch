package drift

import (
	"fmt"
	"sync"
	"time"
)

// WindowPolicy defines the configuration for a sliding time window.
type WindowPolicy struct {
	Size     time.Duration `yaml:"size"`
	MaxItems int           `yaml:"max_items"`
}

// DefaultWindowPolicy returns a WindowPolicy with sensible defaults.
func DefaultWindowPolicy() WindowPolicy {
	return WindowPolicy{
		Size:     5 * time.Minute,
		MaxItems: 100,
	}
}

// Validate returns an error if the policy is misconfigured.
func (p WindowPolicy) Validate() error {
	if p.Size <= 0 {
		return fmt.Errorf("window size must be positive")
	}
	if p.MaxItems <= 0 {
		return fmt.Errorf("window max_items must be positive")
	}
	return nil
}

type windowEntry struct {
	at     time.Time
	result DriftResult
}

// SlidingWindow accumulates DriftResults within a rolling time window.
type SlidingWindow struct {
	mu      sync.Mutex
	policy  WindowPolicy
	entries []windowEntry
}

// NewSlidingWindow creates a SlidingWindow with the given policy.
// If policy is zero-valued, DefaultWindowPolicy is used.
func NewSlidingWindow(p WindowPolicy) (*SlidingWindow, error) {
	if p.Size == 0 && p.MaxItems == 0 {
		p = DefaultWindowPolicy()
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &SlidingWindow{policy: p}, nil
}

// Add inserts a result into the window, evicting stale or excess entries.
func (w *SlidingWindow) Add(r DriftResult) {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-w.policy.Size)
	filtered := w.entries[:0]
	for _, e := range w.entries {
		if e.at.After(cutoff) {
			filtered = append(filtered, e)
		}
	}
	filtered = append(filtered, windowEntry{at: now, result: r})
	if len(filtered) > w.policy.MaxItems {
		filtered = filtered[len(filtered)-w.policy.MaxItems:]
	}
	w.entries = filtered
}

// Results returns a snapshot of all results currently in the window.
func (w *SlidingWindow) Results() []DriftResult {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]DriftResult, len(w.entries))
	for i, e := range w.entries {
		out[i] = e.result
	}
	return out
}

// Len returns the number of entries currently in the window.
func (w *SlidingWindow) Len() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.entries)
}
