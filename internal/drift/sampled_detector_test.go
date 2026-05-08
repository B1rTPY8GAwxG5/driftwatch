package drift

import (
	"testing"
)

// stubDetector is a minimal Detector that records how many times Compare
// was called and returns a fixed result.
type stubDetector struct {
	calls  int
	result DriftResult
}

func (s *stubDetector) Compare(live, spec ServiceSpec) DriftResult {
	s.calls++
	return s.result
}

func TestNewSampledDetector_NotNil(t *testing.T) {
	stub := &stubDetector{}
	sampler := NewSampler(1.0, SampleModeRandom)
	sd := NewSampledDetector(stub, sampler)
	if sd == nil {
		t.Fatal("expected non-nil SampledDetector")
	}
}

func TestSampledDetector_SamplerAccessor(t *testing.T) {
	sampler := NewSampler(0.5, SampleModeRoundRobin)
	sd := NewSampledDetector(&stubDetector{}, sampler)
	if sd.Sampler() != sampler {
		t.Fatal("expected sampler to be returned")
	}
}

func TestSampledDetector_FullRate_DelegatesAlways(t *testing.T) {
	stub := &stubDetector{result: DriftResult{Service: "svc", Entries: []DriftEntry{}}}
	sd := NewSampledDetector(stub, NewSampler(1.0, SampleModeRandom))
	spec := ServiceSpec{Name: "svc"}
	for i := 0; i < 5; i++ {
		sd.Compare(spec, spec)
	}
	if stub.calls != 5 {
		t.Fatalf("expected 5 delegate calls, got %d", stub.calls)
	}
}

func TestSampledDetector_ZeroRate_NeverDelegates(t *testing.T) {
	stub := &stubDetector{}
	sd := NewSampledDetector(stub, NewSampler(0, SampleModeRandom))
	spec := ServiceSpec{Name: "svc"}
	for i := 0; i < 5; i++ {
		res := sd.Compare(spec, spec)
		if !res.Skipped {
			t.Fatalf("call %d: expected Skipped=true", i)
		}
	}
	if stub.calls != 0 {
		t.Fatalf("expected 0 delegate calls, got %d", stub.calls)
	}
}

func TestSampledDetector_SkippedResult_HasServiceName(t *testing.T) {
	stub := &stubDetector{}
	sd := NewSampledDetector(stub, NewSampler(0, SampleModeRandom))
	spec := ServiceSpec{Name: "my-service"}
	res := sd.Compare(spec, spec)
	if res.Service != "my-service" {
		t.Fatalf("expected service name 'my-service', got %q", res.Service)
	}
}

func TestSampledDetector_RoundRobin_SamplesEveryOther(t *testing.T) {
	stub := &stubDetector{result: DriftResult{Service: "svc", Entries: []DriftEntry{}}}
	sd := NewSampledDetector(stub, NewSampler(0.5, SampleModeRoundRobin))
	spec := ServiceSpec{Name: "svc"}
	for i := 0; i < 10; i++ {
		sd.Compare(spec, spec)
	}
	// With step=2 and round-robin, 5 out of 10 calls should delegate.
	if stub.calls != 5 {
		t.Fatalf("expected 5 delegate calls, got %d", stub.calls)
	}
}
