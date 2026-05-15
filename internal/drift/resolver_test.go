package drift

import (
	"testing"
	"time"
)

var (
	earlierTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	laterTime   = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
)

func resolverResult(entries []DriftEntry) DriftResult {
	return DriftResult{Service: "svc", Entries: entries}
}

func TestNewResolver_DefaultsToLatest(t *testing.T) {
	r := NewResolver("bogus")
	if r.Strategy() != ResolveStrategyLatest {
		t.Errorf("expected latest, got %s", r.Strategy())
	}
}

func TestNewResolver_KnownStrategies(t *testing.T) {
	for _, s := range []ResolveStrategy{ResolveStrategyLatest, ResolveStrategyEarliest, ResolveStrategySeverest} {
		r := NewResolver(s)
		if r.Strategy() != s {
			t.Errorf("expected %s, got %s", s, r.Strategy())
		}
	}
}

func TestResolver_Resolve_NoDuplicates(t *testing.T) {
	entries := []DriftEntry{
		{Kind: DriftKindImage, Field: "image", DetectedAt: earlierTime},
		{Kind: DriftKindReplicas, Field: "replicas", DetectedAt: laterTime},
	}
	r := NewResolver(ResolveStrategyLatest)
	out := r.Resolve(resolverResult(entries))
	if len(out.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(out.Entries))
	}
}

func TestResolver_Resolve_Latest_KeepsNewer(t *testing.T) {
	entries := []DriftEntry{
		{Kind: DriftKindImage, Field: "image", Actual: "old", DetectedAt: earlierTime},
		{Kind: DriftKindImage, Field: "image", Actual: "new", DetectedAt: laterTime},
	}
	r := NewResolver(ResolveStrategyLatest)
	out := r.Resolve(resolverResult(entries))
	if len(out.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out.Entries))
	}
	if out.Entries[0].Actual != "new" {
		t.Errorf("expected 'new', got %q", out.Entries[0].Actual)
	}
}

func TestResolver_Resolve_Earliest_KeepsOlder(t *testing.T) {
	entries := []DriftEntry{
		{Kind: DriftKindImage, Field: "image", Actual: "old", DetectedAt: earlierTime},
		{Kind: DriftKindImage, Field: "image", Actual: "new", DetectedAt: laterTime},
	}
	r := NewResolver(ResolveStrategyEarliest)
	out := r.Resolve(resolverResult(entries))
	if len(out.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out.Entries))
	}
	if out.Entries[0].Actual != "old" {
		t.Errorf("expected 'old', got %q", out.Entries[0].Actual)
	}
}

func TestResolver_Resolve_Severest_KeepsHigherRank(t *testing.T) {
	entries := []DriftEntry{
		{Kind: DriftKindEnv, Field: "image", Actual: "env-val", DetectedAt: earlierTime},
		{Kind: DriftKindImage, Field: "image", Actual: "img-val", DetectedAt: laterTime},
	}
	r := NewResolver(ResolveStrategySeverest)
	out := r.Resolve(resolverResult(entries))
	if len(out.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out.Entries))
	}
	if out.Entries[0].Kind != DriftKindImage {
		t.Errorf("expected image kind, got %s", out.Entries[0].Kind)
	}
}

func TestResolver_ResolveAll_AppliesPerResult(t *testing.T) {
	results := []DriftResult{
		resolverResult([]DriftEntry{
			{Kind: DriftKindImage, Field: "image", Actual: "a", DetectedAt: earlierTime},
			{Kind: DriftKindImage, Field: "image", Actual: "b", DetectedAt: laterTime},
		}),
		resolverResult([]DriftEntry{
			{Kind: DriftKindReplicas, Field: "replicas", DetectedAt: earlierTime},
		}),
	}
	r := NewResolver(ResolveStrategyLatest)
	out := r.ResolveAll(results)
	if len(out) != 2 {
		t.Fatalf("expected 2 results, got %d", len(out))
	}
	if len(out[0].Entries) != 1 {
		t.Errorf("expected 1 deduplicated entry in first result, got %d", len(out[0].Entries))
	}
}

func TestLoadResolverStrategy_Valid(t *testing.T) {
	for _, s := range []string{"latest", "earliest", "severest"} {
		_, err := LoadResolverStrategy(s)
		if err != nil {
			t.Errorf("unexpected error for %q: %v", s, err)
		}
	}
}

func TestLoadResolverStrategy_Invalid(t *testing.T) {
	_, err := LoadResolverStrategy("unknown")
	if err == nil {
		t.Error("expected error for unknown strategy")
	}
}
