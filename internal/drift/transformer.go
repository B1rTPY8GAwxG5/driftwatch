package drift

// TransformFn is a function that transforms a DriftResult.
type TransformFn func(DriftResult) DriftResult

// Transformer applies a chain of transformation functions to DriftResults.
type Transformer struct {
	fns []TransformFn
}

// NewTransformer creates a new Transformer with the provided transform functions.
func NewTransformer(fns ...TransformFn) *Transformer {
	return &Transformer{fns: fns}
}

// Add appends a transform function to the chain.
func (t *Transformer) Add(fn TransformFn) {
	if fn != nil {
		t.fns = append(t.fns, fn)
	}
}

// Transform applies all registered transform functions to the given result
// in order, returning the final transformed result.
func (t *Transformer) Transform(result DriftResult) DriftResult {
	for _, fn := range t.fns {
		result = fn(result)
	}
	return result
}

// TransformAll applies Transform to each result in the slice.
func (t *Transformer) TransformAll(results []DriftResult) []DriftResult {
	out := make([]DriftResult, len(results))
	for i, r := range results {
		out[i] = t.Transform(r)
	}
	return out
}

// NormaliseServiceName returns a TransformFn that lower-cases the service name
// in the result's spec for consistent downstream processing.
func NormaliseServiceName() TransformFn {
	return func(r DriftResult) DriftResult {
		r.Spec.Name = strings.ToLower(r.Spec.Name)
		return r
	}
}

// RedactEnvValues returns a TransformFn that replaces the Value field of any
// DriftEntry of kind KindEnv with the string "[redacted]" to avoid leaking
// sensitive environment variable values in reports.
func RedactEnvValues() TransformFn {
	return func(r DriftResult) DriftResult {
		entries := make([]DriftEntry, len(r.Entries))
		for i, e := range r.Entries {
			if e.Kind == KindEnv {
				e.Got = "[redacted]"
				e.Want = "[redacted]"
			}
			entries[i] = e
		}
		r.Entries = entries
		return r
	}
}
