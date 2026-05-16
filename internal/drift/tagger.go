package drift

import "strings"

// TaggerRule maps a condition to a set of tags to apply.
type TaggerRule struct {
	// Kind restricts the rule to a specific DriftKind; empty means any.
	Kind  DriftKind
	Tags  []string
}

// Tagger attaches tags to DriftResult entries based on configurable rules.
type Tagger struct {
	rules      []TaggerRule
	staticTags []string
}

// NewTagger returns a Tagger with optional static tags applied to every result.
func NewTagger(staticTags ...string) *Tagger {
	return &Tagger{staticTags: staticTags}
}

// AddRule registers a TaggerRule with the Tagger.
func (t *Tagger) AddRule(r TaggerRule) {
	if len(r.Tags) == 0 {
		return
	}
	t.rules = append(t.rules, r)
}

// Tag returns the combined set of tags that apply to the given DriftResult.
// Static tags are always included. Rule-based tags are included when the rule
// kind matches at least one entry in the result (or the rule kind is empty).
func (t *Tagger) Tag(result DriftResult) []string {
	seen := make(map[string]struct{})
	var out []string

	add := func(tags []string) {
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			if _, ok := seen[tag]; !ok {
				seen[tag] = struct{}{}
				out = append(out, tag)
			}
		}
	}

	add(t.staticTags)

	for _, rule := range t.rules {
		if rule.Kind == "" {
			add(rule.Tags)
			continue
		}
		for _, entry := range result.Entries {
			if entry.Kind == rule.Kind {
				add(rule.Tags)
				break
			}
		}
	}

	return out
}

// TagAll applies Tag to each result and returns a map of service name to tags.
func (t *Tagger) TagAll(results []DriftResult) map[string][]string {
	out := make(map[string][]string, len(results))
	for _, r := range results {
		out[r.Service] = t.Tag(r)
	}
	return out
}
