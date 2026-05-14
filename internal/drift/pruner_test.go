package drift

import (
	"testing"
	"time"
)

func makePrunerResult(service string, age time.Duration, drifted bool) DriftResult {
	r := DriftResult{
		Service:   service,
		Timestamp: time.Now().Add(-age),
	}
	if drifted {
		r.Entries = []DriftEntry{{Kind: KindImage, Field: "image", Got: "v2", Want: "v1"}}
	}
	return r
}

func TestDefaultPrunePolicy_Values(t *testing.T) {
	p := DefaultPrunePolicy()
	if p.MaxAge != 72*time.Hour {
		t.Errorf("expected MaxAge 72h, got %v", p.MaxAge)
	}
	if p.MaxEntries != 100 {
		t.Errorf("expected MaxEntries 100, got %d", p.MaxEntries)
	}
}

func TestNewPruner_NotNil(t *testing.T) {
	p := NewPruner(DefaultPrunePolicy())
	if p == nil {
		t.Fatal("expected non-nil Pruner")
	}
}

func TestPruner_Prune_RemovesOldResults(t *testing.T) {
	policy := PrunePolicy{MaxAge: time.Hour, MaxEntries: 0}
	pr := NewPruner(policy)

	results := []DriftResult{
		makePrunerResult("svc-a", 2*time.Hour, true),  // too old
		makePrunerResult("svc-a", 30*time.Minute, false), // fresh
	}

	out := pr.Prune(results)
	if len(out) != 1 {
		t.Fatalf("expected 1 result, got %d", len(out))
	}
	if out[0].Service != "svc-a" {
		t.Errorf("unexpected service: %s", out[0].Service)
	}
}

func TestPruner_Prune_KeepsAllWithinAge(t *testing.T) {
	policy := PrunePolicy{MaxAge: time.Hour, MaxEntries: 0}
	pr := NewPruner(policy)

	results := []DriftResult{
		makePrunerResult("svc-b", 10*time.Minute, true),
		makePrunerResult("svc-b", 20*time.Minute, false),
	}

	out := pr.Prune(results)
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
}

func TestPruner_Prune_EnforcesMaxEntries(t *testing.T) {
	policy := PrunePolicy{MaxAge: time.Hour, MaxEntries: 2}
	pr := NewPruner(policy)

	results := []DriftResult{
		makePrunerResult("svc-c", 50*time.Minute, true),
		makePrunerResult("svc-c", 40*time.Minute, true),
		makePrunerResult("svc-c", 30*time.Minute, true),
	}

	out := pr.Prune(results)
	if len(out) != 2 {
		t.Fatalf("expected 2 results after cap, got %d", len(out))
	}
}

func TestPruner_PruneAll_FlattensAndPrunes(t *testing.T) {
	policy := PrunePolicy{MaxAge: time.Hour, MaxEntries: 0}
	pr := NewPruner(policy)

	m := map[string][]DriftResult{
		"svc-x": {
			makePrunerResult("svc-x", 2*time.Hour, true),
			makePrunerResult("svc-x", 5*time.Minute, false),
		},
		"svc-y": {
			makePrunerResult("svc-y", 10*time.Minute, true),
		},
	}

	out := pr.PruneAll(m)
	if len(out) != 2 {
		t.Fatalf("expected 2 surviving results, got %d", len(out))
	}
}

func TestPruner_Prune_EmptyInput(t *testing.T) {
	pr := NewPruner(DefaultPrunePolicy())
	out := pr.Prune(nil)
	if len(out) != 0 {
		t.Errorf("expected empty result, got %d", len(out))
	}
}
