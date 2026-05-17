package drift

import (
	"testing"
	"time"
)

func makeShrinkerResult(service string, drifted bool, age time.Duration) DriftResult {
	entries := []DriftEntry{}
	if drifted {
		entries = append(entries, DriftEntry{Kind: KindImage, Field: "image", Declared: "a", Observed: "b"})
	}
	return DriftResult{
		Service:   service,
		Entries:   entries,
		CheckedAt: time.Now().Add(-age),
	}
}

func TestDefaultShrinkerOptions_Values(t *testing.T) {
	opts := DefaultShrinkerOptions()
	if opts.MaxAge != 24*time.Hour {
		t.Errorf("expected 24h MaxAge, got %v", opts.MaxAge)
	}
	if opts.MaxResults != 500 {
		t.Errorf("expected 500 MaxResults, got %d", opts.MaxResults)
	}
	if !opts.KeepDrifted {
		t.Error("expected KeepDrifted to be true")
	}
}

func TestNewShrinker_NotNil(t *testing.T) {
	s := NewShrinker(DefaultShrinkerOptions())
	if s == nil {
		t.Fatal("expected non-nil Shrinker")
	}
}

func TestNewShrinker_ZeroOptions_UsesDefaults(t *testing.T) {
	s := NewShrinker(ShrinkerOptions{})
	if s.opts.MaxResults != 500 {
		t.Errorf("expected default MaxResults=500, got %d", s.opts.MaxResults)
	}
}

func TestShrinker_Shrink_RemovesOldClean(t *testing.T) {
	s := NewShrinker(ShrinkerOptions{MaxAge: time.Hour, KeepDrifted: false})
	input := []DriftResult{
		makeShrinkerResult("svc-a", false, 2*time.Hour),
		makeShrinkerResult("svc-b", false, 30*time.Minute),
	}
	out := s.Shrink(input)
	if len(out) != 1 || out[0].Service != "svc-b" {
		t.Errorf("expected only svc-b, got %v", out)
	}
}

func TestShrinker_Shrink_KeepsDriftedWhenFlagSet(t *testing.T) {
	s := NewShrinker(ShrinkerOptions{MaxAge: time.Hour, KeepDrifted: true})
	input := []DriftResult{
		makeShrinkerResult("svc-old-drifted", true, 3*time.Hour),
		makeShrinkerResult("svc-old-clean", false, 3*time.Hour),
	}
	out := s.Shrink(input)
	if len(out) != 1 || out[0].Service != "svc-old-drifted" {
		t.Errorf("expected only svc-old-drifted, got %v", out)
	}
}

func TestShrinker_Shrink_CapsMaxResults(t *testing.T) {
	s := NewShrinker(ShrinkerOptions{MaxResults: 2})
	input := []DriftResult{
		makeShrinkerResult("svc-1", false, time.Minute),
		makeShrinkerResult("svc-2", false, time.Minute),
		makeShrinkerResult("svc-3", false, time.Minute),
	}
	out := s.Shrink(input)
	if len(out) != 2 {
		t.Errorf("expected 2 results, got %d", len(out))
	}
}

func TestShrinker_Len_MatchesShrink(t *testing.T) {
	s := NewShrinker(ShrinkerOptions{MaxResults: 3})
	input := []DriftResult{
		makeShrinkerResult("a", false, time.Minute),
		makeShrinkerResult("b", false, time.Minute),
		makeShrinkerResult("c", false, time.Minute),
		makeShrinkerResult("d", false, time.Minute),
	}
	if s.Len(input) != 3 {
		t.Errorf("expected Len=3, got %d", s.Len(input))
	}
}

func TestShrinker_Shrink_EmptyInput(t *testing.T) {
	s := NewShrinker(DefaultShrinkerOptions())
	out := s.Shrink(nil)
	if len(out) != 0 {
		t.Errorf("expected empty output, got %d", len(out))
	}
}
