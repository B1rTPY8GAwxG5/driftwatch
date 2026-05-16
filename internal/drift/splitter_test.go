package drift

import (
	"testing"
)

func TestNewSplitter_DefaultsToService(t *testing.T) {
	s := NewSplitter("bogus")
	if s.Mode() != SplitByService {
		t.Errorf("expected SplitByService, got %q", s.Mode())
	}
}

func TestNewSplitter_KnownModes(t *testing.T) {
	for _, mode := range []SplitMode{SplitByService, SplitByKind, SplitBySeverity} {
		s := NewSplitter(mode)
		if s.Mode() != mode {
			t.Errorf("expected %q, got %q", mode, s.Mode())
		}
	}
}

func TestSplitter_Split_ByService(t *testing.T) {
	results := []DriftResult{
		{Service: "alpha", Entries: []DriftEntry{{Kind: KindImage}}},
		{Service: "beta", Entries: []DriftEntry{{Kind: KindReplicas}}},
		{Service: "alpha", Entries: []DriftEntry{{Kind: KindEnv}}},
	}
	s := NewSplitter(SplitByService)
	buckets := s.Split(results)
	if len(buckets["alpha"]) != 2 {
		t.Errorf("expected 2 results for alpha, got %d", len(buckets["alpha"]))
	}
	if len(buckets["beta"]) != 1 {
		t.Errorf("expected 1 result for beta, got %d", len(buckets["beta"]))
	}
}

func TestSplitter_Split_ByKind(t *testing.T) {
	results := []DriftResult{
		{Service: "svc", Entries: []DriftEntry{{Kind: KindImage}, {Kind: KindReplicas}}},
		{Service: "svc2", Entries: []DriftEntry{{Kind: KindImage}}},
	}
	s := NewSplitter(SplitByKind)
	buckets := s.Split(results)
	if len(buckets[string(KindImage)]) != 2 {
		t.Errorf("expected 2 in image bucket, got %d", len(buckets[string(KindImage)]))
	}
	if len(buckets[string(KindReplicas)]) != 1 {
		t.Errorf("expected 1 in replicas bucket, got %d", len(buckets[string(KindReplicas)]))
	}
}

func TestSplitter_Split_ByKind_NoEntries_FallsToNone(t *testing.T) {
	results := []DriftResult{{Service: "svc", Entries: nil}}
	s := NewSplitter(SplitByKind)
	buckets := s.Split(results)
	if len(buckets["none"]) != 1 {
		t.Errorf("expected 1 in none bucket, got %d", len(buckets["none"]))
	}
}

func TestSplitter_Split_BySeverity(t *testing.T) {
	results := []DriftResult{
		{Service: "svc", Entries: []DriftEntry{{Kind: KindImage}}},
		{Service: "clean", Entries: nil},
	}
	s := NewSplitter(SplitBySeverity)
	buckets := s.Split(results)
	if len(buckets) == 0 {
		t.Error("expected at least one severity bucket")
	}
}

func TestBucketNames_Sorted(t *testing.T) {
	buckets := map[string][]DriftResult{
		"zebra": {},
		"alpha": {},
		"mango": {},
	}
	names := BucketNames(buckets)
	expected := []string{"alpha", "mango", "zebra"}
	for i, n := range names {
		if n != expected[i] {
			t.Errorf("pos %d: expected %q, got %q", i, expected[i], n)
		}
	}
}

func TestSplitter_Split_EmptyService_FallsToUnknown(t *testing.T) {
	results := []DriftResult{{Service: "", Entries: []DriftEntry{{Kind: KindImage}}}}
	s := NewSplitter(SplitByService)
	buckets := s.Split(results)
	if len(buckets["unknown"]) != 1 {
		t.Errorf("expected 1 in unknown bucket, got %d", len(buckets["unknown"]))
	}
}
