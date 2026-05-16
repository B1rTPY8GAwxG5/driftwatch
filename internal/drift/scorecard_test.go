package drift

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func makeScorecardResult(service string, drifted bool) DriftResult {
	r := DriftResult{Service: service}
	if drifted {
		r.Entries = []DriftEntry{{Kind: KindImage, Field: "image", Declared: "v1", Observed: "v2"}}
	}
	return r
}

func TestNewScorecard_Empty(t *testing.T) {
	sc := NewScorecard()
	if sc == nil {
		t.Fatal("expected non-nil scorecard")
	}
	if len(sc.Entries()) != 0 {
		t.Errorf("expected 0 entries, got %d", len(sc.Entries()))
	}
}

func TestScorecard_Add_EmptyServiceIgnored(t *testing.T) {
	sc := NewScorecard()
	sc.Add(ScorecardEntry{Service: "", Score: 80})
	if len(sc.Entries()) != 0 {
		t.Errorf("expected 0 entries, got %d", len(sc.Entries()))
	}
}

func TestScorecard_Add_ValidEntry(t *testing.T) {
	sc := NewScorecard()
	sc.Add(ScorecardEntry{Service: "svc-a", Score: 90, RecordedAt: time.Now()})
	if len(sc.Entries()) != 1 {
		t.Errorf("expected 1 entry, got %d", len(sc.Entries()))
	}
}

func TestScorecard_Entries_SortedDescending(t *testing.T) {
	sc := NewScorecard()
	sc.Add(ScorecardEntry{Service: "svc-low", Score: 40})
	sc.Add(ScorecardEntry{Service: "svc-high", Score: 95})
	sc.Add(ScorecardEntry{Service: "svc-mid", Score: 70})
	entries := sc.Entries()
	if entries[0].Service != "svc-high" {
		t.Errorf("expected svc-high first, got %s", entries[0].Service)
	}
	if entries[2].Service != "svc-low" {
		t.Errorf("expected svc-low last, got %s", entries[2].Service)
	}
}

func TestGrade_AllBands(t *testing.T) {
	cases := []struct{ score int; want string }{
		{100, "A"}, {90, "A"}, {75, "B"}, {60, "C"}, {40, "D"}, {39, "F"}, {0, "F"},
	}
	for _, c := range cases {
		if g := Grade(c.score); g != c.want {
			t.Errorf("Grade(%d) = %q, want %q", c.score, g, c.want)
		}
	}
}

func TestBuildScorecardEntry_NoDrift(t *testing.T) {
	r := makeScorecardResult("svc-clean", false)
	e := BuildScorecardEntry(r, 95)
	if e.Drifted {
		t.Error("expected Drifted=false")
	}
	if e.Grade != "A" {
		t.Errorf("expected grade A, got %s", e.Grade)
	}
	if len(e.Kinds) != 0 {
		t.Errorf("expected no kinds, got %v", e.Kinds)
	}
}

func TestBuildScorecardEntry_Drifted(t *testing.T) {
	r := makeScorecardResult("svc-drift", true)
	e := BuildScorecardEntry(r, 30)
	if !e.Drifted {
		t.Error("expected Drifted=true")
	}
	if e.Grade != "F" {
		t.Errorf("expected grade F, got %s", e.Grade)
	}
	if len(e.Kinds) == 0 {
		t.Error("expected at least one kind")
	}
}

func TestWriteScorecardSummary_Empty(t *testing.T) {
	var buf bytes.Buffer
	WriteScorecardSummary(&buf, NewScorecard())
	if !strings.Contains(buf.String(), "no entries") {
		t.Errorf("expected 'no entries' in output, got %q", buf.String())
	}
}

func TestWriteScorecardSummary_WithEntries(t *testing.T) {
	sc := NewScorecard()
	sc.Add(ScorecardEntry{Service: "my-service", Score: 80, Grade: "B", Drifted: false})
	var buf bytes.Buffer
	WriteScorecardSummary(&buf, sc)
	out := buf.String()
	if !strings.Contains(out, "my-service") {
		t.Errorf("expected service name in output, got %q", out)
	}
	if !strings.Contains(out, "80") {
		t.Errorf("expected score 80 in output, got %q", out)
	}
}
