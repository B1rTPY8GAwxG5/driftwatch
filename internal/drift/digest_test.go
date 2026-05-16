package drift

import (
	"testing"
)

func baseDigestResult(service string) DriftResult {
	return DriftResult{
		Service: service,
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Expected: "nginx:1.24", Actual: "nginx:1.25"},
			{Kind: KindReplicas, Field: "replicas", Expected: "3", Actual: "2"},
			{Kind: KindEnv, Field: "LOG_LEVEL", Expected: "info", Actual: "debug"},
		},
	}
}

func TestNewDigester_NotNil(t *testing.T) {
	d := NewDigester(DefaultDigestOptions())
	if d == nil {
		t.Fatal("expected non-nil Digester")
	}
}

func TestDefaultDigestOptions_AllEnabled(t *testing.T) {
	opts := DefaultDigestOptions()
	if !opts.IncludeEnv || !opts.IncludeReplicas || !opts.IncludeImage {
		t.Error("expected all digest options to be enabled by default")
	}
}

func TestDigester_Compute_NonEmpty(t *testing.T) {
	d := NewDigester(DefaultDigestOptions())
	result := baseDigestResult("svc-a")
	hash := d.Compute(result)
	if hash == "" {
		t.Error("expected non-empty digest")
	}
}

func TestDigester_Compute_Deterministic(t *testing.T) {
	d := NewDigester(DefaultDigestOptions())
	result := baseDigestResult("svc-a")
	h1 := d.Compute(result)
	h2 := d.Compute(result)
	if h1 != h2 {
		t.Errorf("expected deterministic digest, got %q and %q", h1, h2)
	}
}

func TestDigester_Compute_DifferentServices_DifferentDigests(t *testing.T) {
	d := NewDigester(DefaultDigestOptions())
	r1 := baseDigestResult("svc-a")
	r2 := baseDigestResult("svc-b")
	if d.Compute(r1) == d.Compute(r2) {
		t.Error("expected different digests for different services")
	}
}

func TestDigester_Compute_ExcludeEnv_IgnoresEnvEntries(t *testing.T) {
	opts := DefaultDigestOptions()
	opts.IncludeEnv = false
	d := NewDigester(opts)

	r1 := baseDigestResult("svc-a")
	r2 := baseDigestResult("svc-a")
	r2.Entries = append(r2.Entries, DriftEntry{Kind: KindEnv, Field: "EXTRA", Expected: "x", Actual: "y"})

	if d.Compute(r1) != d.Compute(r2) {
		t.Error("expected equal digests when env is excluded")
	}
}

func TestDigester_Compute_ExcludeImage_IgnoresImageEntry(t *testing.T) {
	opts := DefaultDigestOptions()
	opts.IncludeImage = false
	d := NewDigester(opts)

	r1 := DriftResult{Service: "svc-a", Entries: []DriftEntry{}}
	r2 := DriftResult{Service: "svc-a", Entries: []DriftEntry{
		{Kind: KindImage, Field: "image", Expected: "nginx:1.24", Actual: "nginx:1.25"},
	}}

	if d.Compute(r1) != d.Compute(r2) {
		t.Error("expected equal digests when image is excluded")
	}
}

func TestDigester_Compute_EnvOrder_Stable(t *testing.T) {
	d := NewDigester(DefaultDigestOptions())

	r1 := DriftResult{Service: "svc", Entries: []DriftEntry{
		{Kind: KindEnv, Field: "B", Expected: "1", Actual: "2"},
		{Kind: KindEnv, Field: "A", Expected: "x", Actual: "y"},
	}}
	r2 := DriftResult{Service: "svc", Entries: []DriftEntry{
		{Kind: KindEnv, Field: "A", Expected: "x", Actual: "y"},
		{Kind: KindEnv, Field: "B", Expected: "1", Actual: "2"},
	}}

	if d.Compute(r1) != d.Compute(r2) {
		t.Error("expected stable digest regardless of env entry order")
	}
}
