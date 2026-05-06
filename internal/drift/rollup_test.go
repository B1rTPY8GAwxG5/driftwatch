package drift

import (
	"testing"
)

func makeRollupResult(service string, entries []DriftEntry) DriftResult {
	return DriftResult{Service: service, Entries: entries}
}

func TestBuildRollup_Empty(t *testing.T) {
	r := BuildRollup(nil)
	if r.Total != 0 || r.Drifted != 0 {
		t.Fatalf("expected empty rollup, got %+v", r)
	}
	if r.Severity != RollupClean {
		t.Errorf("expected clean severity, got %s", r.Severity)
	}
}

func TestBuildRollup_AllClean(t *testing.T) {
	results := []DriftResult{
		makeRollupResult("svc-a", nil),
		makeRollupResult("svc-b", nil),
	}
	r := BuildRollup(results)
	if r.Total != 2 {
		t.Errorf("expected total 2, got %d", r.Total)
	}
	if r.Drifted != 0 {
		t.Errorf("expected 0 drifted, got %d", r.Drifted)
	}
	if r.Severity != RollupClean {
		t.Errorf("expected clean, got %s", r.Severity)
	}
}

func TestBuildRollup_SomeDrifted(t *testing.T) {
	results := []DriftResult{
		makeRollupResult("svc-a", []DriftEntry{{Kind: KindImage}}),
		makeRollupResult("svc-b", nil),
	}
	r := BuildRollup(results)
	if r.Drifted != 1 {
		t.Errorf("expected 1 drifted, got %d", r.Drifted)
	}
}

func TestBuildRollup_EntriesKinds(t *testing.T) {
	results := []DriftResult{
		makeRollupResult("svc-x", []DriftEntry{{Kind: KindImage}, {Kind: KindReplicas}}),
	}
	r := BuildRollup(results)
	if len(r.Entries) != 1 {
		t.Fatalf("expected 1 entry")
	}
	if len(r.Entries[0].Kinds) != 2 {
		t.Errorf("expected 2 kinds, got %d", len(r.Entries[0].Kinds))
	}
}

func TestRollupSeverity_Levels(t *testing.T) {
	cases := []struct {
		score    int
		expected RollupSeverity
	}{
		{0, RollupClean},
		{10, RollupLow},
		{30, RollupMedium},
		{60, RollupHigh},
		{90, RollupCritical},
	}
	for _, c := range cases {
		got := rollupSeverity(c.score)
		if got != c.expected {
			t.Errorf("score %d: expected %s, got %s", c.score, c.expected, got)
		}
	}
}

func TestRollup_Summary(t *testing.T) {
	r := Rollup{Total: 3, Drifted: 1, Severity: RollupMedium}
	s := r.Summary()
	if s == "" {
		t.Error("expected non-empty summary")
	}
	expected := "services=3 drifted=1 severity=medium"
	if s != expected {
		t.Errorf("expected %q, got %q", expected, s)
	}
}
