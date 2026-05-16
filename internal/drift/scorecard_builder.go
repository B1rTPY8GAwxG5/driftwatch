package drift

// ScorecardBuilder composes a Scorecard from a slice of DriftResults
// using a Scorer to derive numeric scores.
type ScorecardBuilder struct {
	scorer    func(DriftResult) ScoredResult
	scorecard *Scorecard
}

// NewScorecardBuilder returns a ScorecardBuilder backed by the default scorer.
func NewScorecardBuilder() *ScorecardBuilder {
	return &ScorecardBuilder{
		scorer:    ScoreResult,
		scorecard: NewScorecard(),
	}
}

// WithScorer replaces the scoring function used by the builder.
func (b *ScorecardBuilder) WithScorer(fn func(DriftResult) ScoredResult) *ScorecardBuilder {
	if fn != nil {
		b.scorer = fn
	}
	return b
}

// Add scores a DriftResult and appends it to the internal scorecard.
func (b *ScorecardBuilder) Add(r DriftResult) {
	scored := b.scorer(r)
	entry := BuildScorecardEntry(r, scored.Score)
	b.scorecard.Add(entry)
}

// AddAll scores and adds multiple DriftResults.
func (b *ScorecardBuilder) AddAll(results []DriftResult) {
	for _, r := range results {
		b.Add(r)
	}
}

// Build returns the completed Scorecard.
func (b *ScorecardBuilder) Build() *Scorecard {
	return b.scorecard
}

// Reset clears all accumulated entries.
func (b *ScorecardBuilder) Reset() {
	b.scorecard = NewScorecard()
}
