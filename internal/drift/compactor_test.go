package drift

import (
	"testing"
	"time"
)

func makeCompactResult(service string, drifted bool, age time.Duration) DriftResult {
	r := DriftResult{
		Service:   service,
		Timestamp: time.Now().Add(-age),
	}
	if drifted {
		r.Entries = []DriftEntry{{Kind: KindImage, Field: "image", Declared: "a", Observed: "b"}}
	}
	return r
}

func TestDefaultCompactPolicy_Values(t *testing.T) {
	p := DefaultCompactPolicy()
	if p.MaxAge != 24*time.Hour {
		t.Errorf("expected 24h, got %v", p.MaxAge)
	}
	if p.MaxResults != 50 {
		t.Errorf("expected 50, got %d", p.MaxResults)
	}
	if p.KeepDriftedOnly {
		t.Error("expected KeepDriftedOnly false")
	}
}

func TestNewCompactor_NotNil(t *testing.T) {
	c := NewCompactor(CompactPolicy{})
	if c == nil {
		t.Fatal("expected non-nil compactor")
	}
}

func TestNewCompactor_ZeroPolicy_UsesDefaults(t *testing.T) {
	c := NewCompactor(CompactPolicy{})
	def := DefaultCompactPolicy()
	if c.Policy().MaxAge != def.MaxAge {
		t.Errorf("expected default MaxAge %v, got %v", def.MaxAge, c.Policy().MaxAge)
	}
	if c.Policy().MaxResults != def.MaxResults {
		t.Errorf("expected default MaxResults %d, got %d", def.MaxResults, c.Policy().MaxResults)
	}
}

func TestCompactor_Compact_Empty(t *testing.T) {
	c := NewCompactor(CompactPolicy{})
	out := c.Compact(nil)
	if len(out) != 0 {
		t.Errorf("expected empty, got %d", len(out))
	}
}

func TestCompactor_Compact_RemovesOldResults(t *testing.T) {
	c := NewCompactor(CompactPolicy{MaxAge: time.Hour, MaxResults: 100})
	results := []DriftResult{
		makeCompactResult("svc-a", true, 30*time.Minute),
		makeCompactResult("svc-b", false, 2*time.Hour), // too old
	}
	out := c.Compact(results)
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if out[0].Service != "svc-a" {
		t.Errorf("expected svc-a, got %s", out[0].Service)
	}
}

func TestCompactor_Compact_KeepDriftedOnly(t *testing.T) {
	c := NewCompactor(CompactPolicy{MaxAge: time.Hour, MaxResults: 100, KeepDriftedOnly: true})
	results := []DriftResult{
		makeCompactResult("svc-a", true, 5*time.Minute),
		makeCompactResult("svc-b", false, 5*time.Minute),
	}
	out := c.Compact(results)
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if out[0].Service != "svc-a" {
		t.Errorf("expected svc-a, got %s", out[0].Service)
	}
}

func TestCompactor_Compact_TrimsToMaxResults(t *testing.T) {
	c := NewCompactor(CompactPolicy{MaxAge: time.Hour, MaxResults: 2})
	results := []DriftResult{
		makeCompactResult("svc-a", true, 10*time.Minute),
		makeCompactResult("svc-b", true, 20*time.Minute),
		makeCompactResult("svc-c", true, 30*time.Minute),
	}
	out := c.Compact(results)
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
}

func TestCompactor_Compact_SortedNewestFirst(t *testing.T) {
	c := NewCompactor(CompactPolicy{MaxAge: time.Hour, MaxResults: 10})
	results := []DriftResult{
		makeCompactResult("svc-old", true, 40*time.Minute),
		makeCompactResult("svc-new", true, 5*time.Minute),
	}
	out := c.Compact(results)
	if out[0].Service != "svc-new" {
		t.Errorf("expected svc-new first, got %s", out[0].Service)
	}
}
