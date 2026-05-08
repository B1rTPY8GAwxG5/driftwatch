package drift

// SampledDetector wraps a Detector and skips comparisons that are not
// selected by the configured Sampler, returning a clean result instead.
type SampledDetector struct {
	inner   Detector
	sampler *Sampler
}

// Detector is the interface satisfied by types that can compare a live
// service against its declared spec.
type Detector interface {
	Compare(live, spec ServiceSpec) DriftResult
}

// NewSampledDetector wraps the given Detector with sampling logic.
// When ShouldSample returns false the comparison is skipped and a clean
// DriftResult is returned immediately.
func NewSampledDetector(inner Detector, sampler *Sampler) *SampledDetector {
	if inner == nil {
		panic("drift: NewSampledDetector requires a non-nil inner detector")
	}
	if sampler == nil {
		panic("drift: NewSampledDetector requires a non-nil sampler")
	}
	return &SampledDetector{inner: inner, sampler: sampler}
}

// Compare delegates to the inner Detector only when the Sampler allows it.
// Skipped comparisons return a DriftResult with no entries and Skipped=true.
func (sd *SampledDetector) Compare(live, spec ServiceSpec) DriftResult {
	if !sd.sampler.ShouldSample() {
		return DriftResult{
			Service: spec.Name,
			Entries: []DriftEntry{},
			Skipped: true,
		}
	}
	return sd.inner.Compare(live, spec)
}

// Sampler returns the Sampler used by this SampledDetector.
func (sd *SampledDetector) Sampler() *Sampler { return sd.sampler }
