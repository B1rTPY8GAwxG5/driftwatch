package drift

import (
	"testing"
	"time"
)

var (
	t0 = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 = t0.Add(1 * time.Hour)
	t2 = t0.Add(2 * time.Hour)
)

func cleanTrendResult(svc string) DriftResult {
	return DriftResult{Service: svc, Entries: nil}
}

func driftedTrendResult(svc string, n int) DriftResult {
	entries := make([]DriftEntry, n)
	for i := range entries {
		entries[i] = DriftEntry{Kind: KindImage, Field: "image"}
	}
	return DriftResult{Service: svc, Entries: entries}
}

func TestNewTrendAnalyzer_NotNil(t *testing.T) {
	a := NewTrendAnalyzer()
	if a == nil {
		t.Fatal("expected non-nil TrendAnalyzer")
	}
}

func TestTrendAnalyzer_Summarise_NoData(t *testing.T) {
	a := NewTrendAnalyzer()
	_, ok := a.Summarise("svc-a")
	if ok {
		t.Fatal("expected false for unknown service")
	}
}

func TestTrendAnalyzer_Stable(t *testing.T) {
	a := NewTrendAnalyzer()
	a.Record(driftedTrendResult("svc-a", 2), t0)
	a.Record(driftedTrendResult("svc-a", 2), t1)
	s, ok := a.Summarise("svc-a")
	if !ok {
		t.Fatal("expected summary")
	}
	if s.Direction != TrendStable {
		t.Errorf("expected stable, got %s", s.Direction)
	}
	if s.Delta != 0 {
		t.Errorf("expected delta 0, got %d", s.Delta)
	}
}

func TestTrendAnalyzer_Increasing(t *testing.T) {
	a := NewTrendAnalyzer()
	a.Record(driftedTrendResult("svc-b", 1), t0)
	a.Record(driftedTrendResult("svc-b", 3), t1)
	s, ok := a.Summarise("svc-b")
	if !ok {
		t.Fatal("expected summary")
	}
	if s.Direction != TrendIncreasing {
		t.Errorf("expected increasing, got %s", s.Direction)
	}
	if s.Delta != 2 {
		t.Errorf("expected delta 2, got %d", s.Delta)
	}
}

func TestTrendAnalyzer_Decreasing(t *testing.T) {
	a := NewTrendAnalyzer()
	a.Record(driftedTrendResult("svc-c", 4), t0)
	a.Record(cleanTrendResult("svc-c"), t1)
	s, ok := a.Summarise("svc-c")
	if !ok {
		t.Fatal("expected summary")
	}
	if s.Direction != TrendDecreasing {
		t.Errorf("expected decreasing, got %s", s.Direction)
	}
	if s.Delta != -4 {
		t.Errorf("expected delta -4, got %d", s.Delta)
	}
}

func TestTrendAnalyzer_PointsAreSorted(t *testing.T) {
	a := NewTrendAnalyzer()
	a.Record(driftedTrendResult("svc-d", 1), t2)
	a.Record(driftedTrendResult("svc-d", 2), t0)
	a.Record(driftedTrendResult("svc-d", 3), t1)
	s, _ := a.Summarise("svc-d")
	for i := 1; i < len(s.Points); i++ {
		if s.Points[i].Timestamp.Before(s.Points[i-1].Timestamp) {
			t.Errorf("points not sorted at index %d", i)
		}
	}
}

func TestTrendAnalyzer_Services(t *testing.T) {
	a := NewTrendAnalyzer()
	a.Record(cleanTrendResult("alpha"), t0)
	a.Record(cleanTrendResult("beta"), t0)
	svcs := a.Services()
	if len(svcs) != 2 {
		t.Fatalf("expected 2 services, got %d", len(svcs))
	}
	if svcs[0] != "alpha" || svcs[1] != "beta" {
		t.Errorf("unexpected services: %v", svcs)
	}
}

func TestTrendSummary_String(t *testing.T) {
	s := TrendSummary{Service: "svc-x", Direction: TrendStable, Delta: 0, Points: []TrendPoint{{}}}
	str := s.String()
	if str == "" {
		t.Error("expected non-empty string")
	}
}
