package drift

import "sort"

// SplitMode controls how results are partitioned.
type SplitMode string

const (
	SplitByService SplitMode = "service"
	SplitByKind    SplitMode = "kind"
	SplitBySeverity SplitMode = "severity"
)

// Splitter partitions a slice of DriftResults into named buckets.
type Splitter struct {
	mode SplitMode
}

// NewSplitter returns a Splitter using the given mode.
// Falls back to SplitByService for unrecognised modes.
func NewSplitter(mode SplitMode) *Splitter {
	switch mode {
	case SplitByService, SplitByKind, SplitBySeverity:
		return &Splitter{mode: mode}
	default:
		return &Splitter{mode: SplitByService}
	}
}

// Mode returns the active split mode.
func (s *Splitter) Mode() SplitMode { return s.mode }

// Split partitions results into named buckets according to the splitter mode.
func (s *Splitter) Split(results []DriftResult) map[string][]DriftResult {
	buckets := make(map[string][]DriftResult)
	for _, r := range results {
		keys := s.keysFor(r)
		for _, k := range keys {
			buckets[k] = append(buckets[k], r)
		}
	}
	return buckets
}

// BucketNames returns sorted bucket names from a Split result.
func BucketNames(buckets map[string][]DriftResult) []string {
	names := make([]string, 0, len(buckets))
	for k := range buckets {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func (s *Splitter) keysFor(r DriftResult) []string {
	switch s.mode {
	case SplitByKind:
		seen := make(map[string]struct{})
		var keys []string
		for _, e := range r.Entries {
			k := string(e.Kind)
			if _, ok := seen[k]; !ok {
				seen[k] = struct{}{}
				keys = append(keys, k)
			}
		}
		if len(keys) == 0 {
			return []string{"none"}
		}
		return keys
	case SplitBySeverity:
		_, level := ScoreResult(r)
		return []string{string(level)}
	default: // SplitByService
		if r.Service == "" {
			return []string{"unknown"}
		}
		return []string{r.Service}
	}
}
