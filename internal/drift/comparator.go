package drift

import "fmt"

// ComparatorOption configures a Comparator.
type ComparatorOption func(*Comparator)

// Comparator wraps a Detector with optional filtering, suppression, and policy
// enforcement to produce an enriched DriftResult.
type Comparator struct {
	detector   *Detector
	filter     *Filter
	suppressor *SuppressionStore
	policy     *Policy
}

// NewComparator constructs a Comparator with the given Detector and options.
func NewComparator(d *Detector, opts ...ComparatorOption) *Comparator {
	c := &Comparator{detector: d}
	for _, o := range opts {
		o(c)
	}
	return c
}

// WithFilter attaches a drift filter to the comparator.
func WithFilter(f *Filter) ComparatorOption {
	return func(c *Comparator) { c.filter = f }
}

// WithSuppression attaches a suppression store to the comparator.
func WithSuppression(s *SuppressionStore) ComparatorOption {
	return func(c *Comparator) { c.suppressor = s }
}

// WithPolicy attaches a policy to the comparator.
func WithPolicy(p *Policy) ComparatorOption {
	return func(c *Comparator) { c.policy = p }
}

// Compare runs drift detection on spec vs live, then applies filter,
// suppression, and policy in sequence.
func (c *Comparator) Compare(spec, live ServiceSpec) (DriftResult, error) {
	result, err := c.detector.Compare(spec, live)
	if err != nil {
		return DriftResult{}, fmt.Errorf("comparator: detect: %w", err)
	}

	if c.filter != nil {
		result.Entries = ApplyAll([]Filter{*c.filter}, result.Entries)
	}

	if c.suppressor != nil {
		var kept []DriftEntry
		for _, e := range result.Entries {
			if !IsSuppressed(c.suppressor, spec.Name, e.Kind) {
				kept = append(kept, e)
			}
		}
		result.Entries = kept
	}

	if c.policy != nil {
		result = c.policy.Evaluate(result)
	}

	return result, nil
}
