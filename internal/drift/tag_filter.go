package drift

// TagFilter selects or excludes DriftResults based on label tags attached
// by a Labeler. Only results whose labels satisfy ALL required tags are kept.
type TagFilter struct {
	required map[string]string
	excluded map[string]string
}

// NewTagFilter creates a TagFilter.
// required: label key/value pairs that must ALL be present.
// excluded: label key/value pairs where ANY match causes rejection.
func NewTagFilter(required, excluded map[string]string) *TagFilter {
	r := make(map[string]string, len(required))
	for k, v := range required {
		r[k] = v
	}
	e := make(map[string]string, len(excluded))
	for k, v := range excluded {
		e[k] = v
	}
	return &TagFilter{required: r, excluded: e}
}

// Match returns true when the result's labels satisfy the filter criteria.
func (tf *TagFilter) Match(labels map[string]string) bool {
	for k, v := range tf.required {
		if labels[k] != v {
			return false
		}
	}
	for k, v := range tf.excluded {
		if labels[k] == v {
			return false
		}
	}
	return true
}

// Apply filters a slice of LabeledResults, returning only those that match.
func (tf *TagFilter) Apply(results []LabeledResult) []LabeledResult {
	out := make([]LabeledResult, 0, len(results))
	for _, r := range results {
		if tf.Match(r.Labels) {
			out = append(out, r)
		}
	}
	return out
}

// LabeledResult pairs a DriftResult with its computed labels.
type LabeledResult struct {
	Result DriftResult
	Labels map[string]string
}
