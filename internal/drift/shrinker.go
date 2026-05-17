package drift

import "time"

// ShrinkerOptions controls how a Shrinker compacts a slice of DriftResults.
type ShrinkerOptions struct {
	// MaxAge removes results older than this duration. Zero means no age limit.
	MaxAge time.Duration
	// MaxResults caps the total number of results kept (most-recent first).
	// Zero means no cap.
	MaxResults int
	// KeepDrifted, when true, always retains drifted results regardless of age.
	KeepDrifted bool
}

// DefaultShrinkerOptions returns sensible defaults.
func DefaultShrinkerOptions() ShrinkerOptions {
	return ShrinkerOptions{
		MaxAge:      24 * time.Hour,
		MaxResults:  500,
		KeepDrifted: true,
	}
}

// Shrinker reduces a collection of DriftResults according to configured policy.
type Shrinker struct {
	opts ShrinkerOptions
	now  func() time.Time
}

// NewShrinker creates a Shrinker with the given options.
// Zero-value options fall back to DefaultShrinkerOptions.
func NewShrinker(opts ShrinkerOptions) *Shrinker {
	if opts.MaxAge == 0 && opts.MaxResults == 0 {
		opts = DefaultShrinkerOptions()
	}
	return &Shrinker{opts: opts, now: time.Now}
}

// Shrink applies the configured policy to results and returns the survivors.
// Input order is preserved; oldest entries are removed first when capping.
func (s *Shrinker) Shrink(results []DriftResult) []DriftResult {
	cutoff := s.now().Add(-s.opts.MaxAge)

	filtered := results[:0:len(results)]
	for _, r := range results {
		if s.opts.MaxAge > 0 && r.CheckedAt.Before(cutoff) {
			if s.opts.KeepDrifted && r.HasDrift() {
				filtered = append(filtered, r)
			}
			continue
		}
		filtered = append(filtered, r)
	}

	if s.opts.MaxResults > 0 && len(filtered) > s.opts.MaxResults {
		filtered = filtered[len(filtered)-s.opts.MaxResults:]
	}
	return filtered
}

// Len returns the number of results that would survive shrinking.
func (s *Shrinker) Len(results []DriftResult) int {
	return len(s.Shrink(results))
}
