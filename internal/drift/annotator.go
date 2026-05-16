package drift

import "time"

// AnnotationKey is a typed string for annotation map keys.
type AnnotationKey = string

// Annotator attaches structured metadata annotations to DriftResults.
type Annotator struct {
	static map[AnnotationKey]string
	providers []AnnotationProvider
}

// AnnotationProvider is a function that derives annotations from a DriftResult.
type AnnotationProvider func(r DriftResult) map[AnnotationKey]string

// AnnotatedResult wraps a DriftResult with additional annotations.
type AnnotatedResult struct {
	DriftResult
	Annotations map[AnnotationKey]string `json:"annotations"`
	AnnotatedAt time.Time               `json:"annotated_at"`
}

// NewAnnotator creates an Annotator with optional static key/value pairs.
func NewAnnotator(static map[string]string) *Annotator {
	if static == nil {
		static = make(map[string]string)
	}
	return &Annotator{static: static}
}

// AddProvider registers a dynamic annotation provider.
func (a *Annotator) AddProvider(p AnnotationProvider) {
	if p != nil {
		a.providers = append(a.providers, p)
	}
}

// Annotate applies all static and dynamic annotations to the given result.
func (a *Annotator) Annotate(r DriftResult) AnnotatedResult {
	annotations := make(map[string]string, len(a.static))
	for k, v := range a.static {
		annotations[k] = v
	}
	for _, p := range a.providers {
		for k, v := range p(r) {
			annotations[k] = v
		}
	}
	return AnnotatedResult{
		DriftResult:  r,
		Annotations:  annotations,
		AnnotatedAt:  time.Now().UTC(),
	}
}

// AnnotateAll annotates a slice of DriftResults.
func (a *Annotator) AnnotateAll(results []DriftResult) []AnnotatedResult {
	out := make([]AnnotatedResult, len(results))
	for i, r := range results {
		out[i] = a.Annotate(r)
	}
	return out
}
