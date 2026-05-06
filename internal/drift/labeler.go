package drift

import "fmt"

// LabelSet is a map of string key-value labels attached to a drift result.
type LabelSet map[string]string

// Labeler attaches metadata labels to DriftResults based on configurable rules.
type Labeler struct {
	static map[string]string
	rules  []labelRule
}

type labelRule struct {
	kind  DriftKind
	key   string
	value string
}

// NewLabeler returns a Labeler with the given static labels applied to every result.
func NewLabeler(static map[string]string) *Labeler {
	if static == nil {
		static = make(map[string]string)
	}
	return &Labeler{static: static}
}

// AddKindRule registers a label key/value to apply when a result contains the given DriftKind.
func (l *Labeler) AddKindRule(kind DriftKind, key, value string) {
	l.rules = append(l.rules, labelRule{kind: kind, key: key, value: value})
}

// Label returns a LabelSet for the given DriftResult, merging static labels
// with any rule-based labels that match entries in the result.
func (l *Labeler) Label(result DriftResult) LabelSet {
	out := make(LabelSet)
	for k, v := range l.static {
		out[k] = v
	}
	out["service"] = result.Service
	if result.HasDrift() {
		out["drifted"] = "true"
	} else {
		out["drifted"] = "false"
	}
	kinds := entryKindSet(result.Entries)
	for _, rule := range l.rules {
		if kinds[rule.kind] {
			out[rule.key] = rule.value
		}
	}
	return out
}

// String returns a human-readable representation of the LabelSet.
func (ls LabelSet) String() string {
	if len(ls) == 0 {
		return "{}"
	}
	s := "{"
	for k, v := range ls {
		s += fmt.Sprintf("%s=%q ", k, v)
	}
	return s[:len(s)-1] + "}"
}

func entryKindSet(entries []DriftEntry) map[DriftKind]bool {
	set := make(map[DriftKind]bool, len(entries))
	for _, e := range entries {
		set[e.Kind] = true
	}
	return set
}
