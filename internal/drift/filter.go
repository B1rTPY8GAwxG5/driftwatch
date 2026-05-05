package drift

// FilterFunc is a predicate that returns true if a DriftEntry should be included.
type FilterFunc func(entry DriftEntry) bool

// Filter holds configuration for filtering drift results.
type Filter struct {
	kinds map[DriftKind]struct{}
}

// NewFilter creates a Filter that accepts only the specified DriftKinds.
// If no kinds are provided, all kinds are accepted.
func NewFilter(kinds ...DriftKind) *Filter {
	f := &Filter{kinds: make(map[DriftKind]struct{}, len(kinds))}
	for _, k := range kinds {
		f.kinds[k] = struct{}{}
	}
	return f
}

// Match returns true if the entry passes the filter.
func (f *Filter) Match(entry DriftEntry) bool {
	if len(f.kinds) == 0 {
		return true
	}
	_, ok := f.kinds[entry.Kind]
	return ok
}

// Apply returns a new DriftResult containing only entries that pass the filter.
func (f *Filter) Apply(result DriftResult) DriftResult {
	filtered := make([]DriftEntry, 0, len(result.Entries))
	for _, e := range result.Entries {
		if f.Match(e) {
			filtered = append(filtered, e)
		}
	}
	return DriftResult{
		Service: result.Service,
		Entries: filtered,
	}
}

// ApplyAll filters a slice of DriftResults, removing empty results when dropEmpty is true.
func ApplyAll(f *Filter, results []DriftResult, dropEmpty bool) []DriftResult {
	out := make([]DriftResult, 0, len(results))
	for _, r := range results {
		filtered := f.Apply(r)
		if dropEmpty && !filtered.HasDrift() {
			continue
		}
		out = append(out, filtered)
	}
	return out
}
