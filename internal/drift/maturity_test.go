package drift

import (
	"testing"
)

func TestNewMaturityModel_NotNil(t *testing.T) {
	m := NewMaturityModel()
	if m == nil {
		t.Fatal("expected non-nil MaturityModel")
	}
}

func TestMaturityLevel_String(t *testing.T) {
	cases := []struct {
		level MaturityLevel
		want  string
	}{
		{MaturityUnknown, "unknown"},
		{MaturityUnstable, "unstable"},
		{MaturityDeveloping, "developing"},
		{MaturityStable, "stable"},
		{MaturityMature, "mature"},
	}
	for _, tc := range cases {
		if got := tc.level.String(); got != tc.want {
			t.Errorf("level %d: got %q, want %q", tc.level, got, tc.want)
		}
	}
}

func TestMaturityModel_Evaluate_NoObservations(t *testing.T) {
	m := NewMaturityModel()
	_, err := m.Evaluate("svc-a")
	if err == nil {
		t.Fatal("expected error for unknown service")
	}
}

func TestMaturityModel_Evaluate_AllDrifted_Unstable(t *testing.T) {
	m := NewMaturityModel()
	for i := 0; i < 10; i++ {
		m.Record("svc", true)
	}
	rec, err := m.Evaluate("svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Level != MaturityUnstable {
		t.Errorf("got %v, want Unstable", rec.Level)
	}
	if rec.DriftRate != 1.0 {
		t.Errorf("drift rate: got %v, want 1.0", rec.DriftRate)
	}
}

func TestMaturityModel_Evaluate_NoDrift_Mature(t *testing.T) {
	m := NewMaturityModel()
	for i := 0; i < 20; i++ {
		m.Record("svc", false)
	}
	rec, err := m.Evaluate("svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Level != MaturityMature {
		t.Errorf("got %v, want Mature", rec.Level)
	}
}

func TestMaturityModel_Evaluate_LowDrift_Stable(t *testing.T) {
	m := NewMaturityModel()
	m.Record("svc", true)
	for i := 0; i < 19; i++ {
		m.Record("svc", false)
	}
	rec, err := m.Evaluate("svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Level != MaturityStable {
		t.Errorf("got %v, want Stable", rec.Level)
	}
	if rec.Observed != 20 {
		t.Errorf("observed: got %d, want 20", rec.Observed)
	}
}

func TestMaturityModel_Services_ReturnsRecorded(t *testing.T) {
	m := NewMaturityModel()
	m.Record("alpha", false)
	m.Record("beta", true)
	svcs := m.Services()
	if len(svcs) != 2 {
		t.Errorf("got %d services, want 2", len(svcs))
	}
}
