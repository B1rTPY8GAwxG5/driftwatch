package drift_test

import (
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
)

func driftResult(service string, entries ...drift.DriftEntry) drift.DriftResult {
	return drift.DriftResult{Service: service, Entries: entries}
}

func entry(kind drift.DriftKind) drift.DriftEntry {
	return drift.DriftEntry{Kind: kind, Field: "f", Expected: "a", Actual: "b"}
}

func TestNewFilter_NoKinds_MatchesAll(t *testing.T) {
	f := drift.NewFilter()
	if !f.Match(entry(drift.KindImage)) {
		t.Error("expected filter with no kinds to match all entries")
	}
}

func TestFilter_Match_IncludedKind(t *testing.T) {
	f := drift.NewFilter(drift.KindImage)
	if !f.Match(entry(drift.KindImage)) {
		t.Error("expected KindImage to match")
	}
}

func TestFilter_Match_ExcludedKind(t *testing.T) {
	f := drift.NewFilter(drift.KindImage)
	if f.Match(entry(drift.KindReplicas)) {
		t.Error("expected KindReplicas to not match")
	}
}

func TestFilter_Apply_FiltersEntries(t *testing.T) {
	f := drift.NewFilter(drift.KindImage)
	r := driftResult("svc", entry(drift.KindImage), entry(drift.KindReplicas))
	out := f.Apply(r)
	if len(out.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out.Entries))
	}
	if out.Entries[0].Kind != drift.KindImage {
		t.Errorf("expected KindImage, got %v", out.Entries[0].Kind)
	}
}

func TestFilter_Apply_PreservesService(t *testing.T) {
	f := drift.NewFilter(drift.KindImage)
	r := driftResult("my-service", entry(drift.KindImage))
	out := f.Apply(r)
	if out.Service != "my-service" {
		t.Errorf("expected service 'my-service', got %q", out.Service)
	}
}

func TestApplyAll_DropEmpty(t *testing.T) {
	f := drift.NewFilter(drift.KindImage)
	results := []drift.DriftResult{
		driftResult("svc-a", entry(drift.KindImage)),
		driftResult("svc-b", entry(drift.KindReplicas)),
	}
	out := drift.ApplyAll(f, results, true)
	if len(out) != 1 {
		t.Fatalf("expected 1 result after dropEmpty, got %d", len(out))
	}
	if out[0].Service != "svc-a" {
		t.Errorf("unexpected service %q", out[0].Service)
	}
}

func TestApplyAll_KeepEmpty(t *testing.T) {
	f := drift.NewFilter(drift.KindImage)
	results := []drift.DriftResult{
		driftResult("svc-a", entry(drift.KindReplicas)),
	}
	out := drift.ApplyAll(f, results, false)
	if len(out) != 1 {
		t.Fatalf("expected 1 result when not dropping empty, got %d", len(out))
	}
	if out[0].HasDrift() {
		t.Error("expected no drift after filtering")
	}
}
