package drift

import (
	"testing"
)

var fingerprintResult = DriftResult{
	Service: "api",
	Entries: []DriftEntry{
		{Kind: KindImage, Field: "image", Declared: "nginx:1.24", Detected: "nginx:1.25"},
		{Kind: KindReplicas, Field: "replicas", Declared: "3", Detected: "2"},
	},
}

func TestNewFingerprinter_NotNil(t *testing.T) {
	fp := NewFingerprinter(DefaultFingerprintOptions())
	if fp == nil {
		t.Fatal("expected non-nil Fingerprinter")
	}
}

func TestFingerprinter_Compute_NonEmpty(t *testing.T) {
	fp := NewFingerprinter(DefaultFingerprintOptions())
	result := fp.Compute(fingerprintResult)
	if result == "" {
		t.Fatal("expected non-empty fingerprint")
	}
}

func TestFingerprinter_Compute_Deterministic(t *testing.T) {
	fp := NewFingerprinter(DefaultFingerprintOptions())
	a := fp.Compute(fingerprintResult)
	b := fp.Compute(fingerprintResult)
	if !a.Equal(b) {
		t.Fatalf("expected deterministic fingerprint, got %s and %s", a, b)
	}
}

func TestFingerprinter_Compute_DifferentResults_DifferentFingerprints(t *testing.T) {
	fp := NewFingerprinter(DefaultFingerprintOptions())
	other := DriftResult{
		Service: "worker",
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "redis:6", Detected: "redis:7"},
		},
	}
	a := fp.Compute(fingerprintResult)
	b := fp.Compute(other)
	if a.Equal(b) {
		t.Fatal("expected different fingerprints for different results")
	}
}

func TestFingerprinter_Compute_ExcludeEnv(t *testing.T) {
	opts := FingerprintOptions{IncludeImage: true, IncludeReplicas: true, IncludeEnv: false}
	fp := NewFingerprinter(opts)

	withEnv := DriftResult{
		Service: "svc",
		Entries: []DriftEntry{
			{Kind: KindEnv, Field: "SECRET", Declared: "a", Detected: "b"},
		},
	}
	withoutEnv := DriftResult{Service: "svc", Entries: nil}

	a := fp.Compute(withEnv)
	b := fp.Compute(withoutEnv)
	if !a.Equal(b) {
		t.Fatalf("expected equal fingerprints when env excluded, got %s and %s", a, b)
	}
}

func TestFingerprint_String(t *testing.T) {
	fp := Fingerprint("abcd1234")
	if fp.String() != "abcd1234" {
		t.Fatalf("unexpected string: %s", fp.String())
	}
}

func TestFingerprinter_Compute_StableAcrossEntryOrder(t *testing.T) {
	fp := NewFingerprinter(DefaultFingerprintOptions())

	r1 := DriftResult{
		Service: "svc",
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "img:1", Detected: "img:2"},
			{Kind: KindReplicas, Field: "replicas", Declared: "2", Detected: "1"},
		},
	}
	r2 := DriftResult{
		Service: "svc",
		Entries: []DriftEntry{
			{Kind: KindReplicas, Field: "replicas", Declared: "2", Detected: "1"},
			{Kind: KindImage, Field: "image", Declared: "img:1", Detected: "img:2"},
		},
	}

	if !fp.Compute(r1).Equal(fp.Compute(r2)) {
		t.Fatal("expected same fingerprint regardless of entry order")
	}
}
