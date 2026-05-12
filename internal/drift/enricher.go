package drift

import "time"

// EnrichmentContext holds additional metadata to attach to a DriftResult.
type EnrichmentContext struct {
	Environment string
	Cluster     string
	Region      string
	Annotations map[string]string
}

// EnrichedResult wraps a DriftResult with contextual metadata.
type EnrichedResult struct {
	Result    DriftResult
	Context   EnrichmentContext
	EnrichedAt time.Time
}

// Enricher attaches contextual metadata to drift results.
type Enricher struct {
	ctx EnrichmentContext
}

// NewEnricher creates an Enricher with the given context.
func NewEnricher(ctx EnrichmentContext) *Enricher {
	return &Enricher{ctx: ctx}
}

// Enrich wraps a DriftResult with the stored context and a timestamp.
func (e *Enricher) Enrich(r DriftResult) EnrichedResult {
	return EnrichedResult{
		Result:     r,
		Context:    e.ctx,
		EnrichedAt: time.Now().UTC(),
	}
}

// EnrichAll enriches a slice of DriftResults.
func (e *Enricher) EnrichAll(results []DriftResult) []EnrichedResult {
	out := make([]EnrichedResult, 0, len(results))
	for _, r := range results {
		out = append(out, e.Enrich(r))
	}
	return out
}

// HasAnnotation reports whether the enrichment context contains the given key.
func (er EnrichedResult) HasAnnotation(key string) bool {
	_, ok := er.Context.Annotations[key]
	return ok
}
