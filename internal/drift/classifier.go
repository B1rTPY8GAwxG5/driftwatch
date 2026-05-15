package drift

import "strings"

// ClassifierRule maps a drift kind and optional field pattern to a category label.
type ClassifierRule struct {
	Kind     DriftKind
	Field    string // empty matches any field
	Category string
}

// ClassifiedResult wraps a DriftResult with per-entry category annotations.
type ClassifiedResult struct {
	DriftResult
	Categories map[string]string // entry key (kind:field) -> category
}

// Classifier assigns category labels to drift entries based on configurable rules.
type Classifier struct {
	rules []ClassifierRule
}

// NewClassifier returns a Classifier with the provided rules.
func NewClassifier(rules []ClassifierRule) *Classifier {
	return &Classifier{rules: rules}
}

// Classify annotates a DriftResult with category labels derived from matching rules.
// Entries that match no rule receive the category "uncategorised".
func (c *Classifier) Classify(result DriftResult) ClassifiedResult {
	cats := make(map[string]string, len(result.Entries))
	for _, e := range result.Entries {
		key := entryKey(e)
		cats[key] = c.categoryFor(e)
	}
	return ClassifiedResult{
		DriftResult: result,
		Categories:  cats,
	}
}

// ClassifyAll classifies a slice of DriftResults.
func (c *Classifier) ClassifyAll(results []DriftResult) []ClassifiedResult {
	out := make([]ClassifiedResult, len(results))
	for i, r := range results {
		out[i] = c.Classify(r)
	}
	return out
}

func (c *Classifier) categoryFor(e DriftEntry) string {
	for _, r := range c.rules {
		if r.Kind != e.Kind {
			continue
		}
		if r.Field == "" || strings.EqualFold(r.Field, e.Field) {
			return r.Category
		}
	}
	return "uncategorised"
}

func entryKey(e DriftEntry) string {
	return string(e.Kind) + ":" + e.Field
}
