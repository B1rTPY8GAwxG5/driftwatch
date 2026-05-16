package drift

import (
	"testing"
)

func TestNewScorecardBuilder_NotNil(t *testing.T) {
	b := NewScorecardBuilder()
	if b == nil {
		t.Fatal("expected non-nil builder")
	}
}

func TestScorecardBuilder_Add_CleanResult(t *testing.T) {
	b := NewScorecardBuilder()
	r := makeScorecardResult("svc-clean", false)
	b.Add(r)
	entries := b.Build().Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Service != "svc-clean" {
		t.Errorf("unexpected service: %s", entries[0].Service)
	}
}

func TestScorecardBuilder_Add_DriftedResult_LowerScore(t *testing.T) {
	b := NewScorecardBuilder()
	clean := makeScorecardResult("svc-clean", false)
	drifted := makeScorecardResult("svc-drifted", true)
	b.Add(clean)
	b.Add(drifted)
	entries := b.Build().Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// Sorted descending: clean should score higher.
	if entries[0].Service != "svc-clean" {
		t.Errorf("expected svc-clean first, got %s", entries[0].Service)
	}
}

func TestScorecardBuilder_AddAll(t *testing.T) {
	b := NewScorecardBuilder()
	results := []DriftResult{
		makeScorecardResult("svc-a", false),
		makeScorecardResult("svc-b", true),
		makeScorecardResult("svc-c", false),
	}
	b.AddAll(results)
	if len(b.Build().Entries()) != 3 {
		t.Errorf("expected 3 entries")
	}
}

func TestScorecardBuilder_Reset_ClearsEntries(t *testing.T) {
	b := NewScorecardBuilder()
	b.Add(makeScorecardResult("svc-a", false))
	b.Reset()
	if len(b.Build().Entries()) != 0 {
		t.Errorf("expected 0 entries after reset")
	}
}

func TestScorecardBuilder_WithScorer_CustomFn(t *testing.T) {
	customScorer := func(_ DriftResult) ScoredResult {
		return ScoredResult{Score: 42}
	}
	b := NewScorecardBuilder().WithScorer(customScorer)
	b.Add(makeScorecardResult("svc-x", true))
	entries := b.Build().Entries()
	if entries[0].Score != 42 {
		t.Errorf("expected score 42, got %d", entries[0].Score)
	}
	if entries[0].Grade != "D" {
		t.Errorf("expected grade D for score 42, got %s", entries[0].Grade)
	}
}

func TestScorecardBuilder_WithScorer_NilIgnored(t *testing.T) {
	b := NewScorecardBuilder()
	b.WithScorer(nil)
	b.Add(makeScorecardResult("svc-y", false))
	// Should not panic; default scorer used.
	if len(b.Build().Entries()) != 1 {
		t.Error("expected 1 entry")
	}
}
