package drift

import "strings"

// RedactorFunc is a function that redacts a sensitive value.
type RedactorFunc func(key, value string) string

// Redactor masks sensitive fields in DriftResult entries before they are
// forwarded to exporters, notifiers, or audit logs.
type Redactor struct {
	sensitiveKeys []string
	mask          string
	fn            RedactorFunc
}

// NewRedactor returns a Redactor that replaces values whose keys contain any
// of the provided sensitiveKeys (case-insensitive) with mask.
func NewRedactor(sensitiveKeys []string, mask string) *Redactor {
	if mask == "" {
		mask = "***"
	}
	return &Redactor{
		sensitiveKeys: sensitiveKeys,
		mask:          mask,
	}
}

// WithFunc replaces the default key-matching logic with a custom RedactorFunc.
func (r *Redactor) WithFunc(fn RedactorFunc) *Redactor {
	r.fn = fn
	return r
}

// Redact returns a copy of result with sensitive DriftEntry values masked.
func (r *Redactor) Redact(result DriftResult) DriftResult {
	redacted := make([]DriftEntry, len(result.Entries))
	for i, e := range result.Entries {
		redacted[i] = r.redactEntry(e)
	}
	return DriftResult{
		Service: result.Service,
		Entries: redacted,
	}
}

// RedactAll applies Redact to every result in the slice.
func (r *Redactor) RedactAll(results []DriftResult) []DriftResult {
	out := make([]DriftResult, len(results))
	for i, res := range results {
		out[i] = r.Redact(res)
	}
	return out
}

func (r *Redactor) redactEntry(e DriftEntry) DriftEntry {
	if r.fn != nil {
		e.Got = r.fn(e.Field, e.Got)
		e.Want = r.fn(e.Field, e.Want)
		return e
	}
	if r.isSensitive(e.Field) {
		e.Got = r.mask
		e.Want = r.mask
	}
	return e
}

func (r *Redactor) isSensitive(field string) bool {
	lower := strings.ToLower(field)
	for _, k := range r.sensitiveKeys {
		if strings.Contains(lower, strings.ToLower(k)) {
			return true
		}
	}
	return false
}
