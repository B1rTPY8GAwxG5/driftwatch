package drift

import (
	"testing"
	"time"
)

func escalatedResult(service string) DriftResult {
	return DriftResult{
		Service: service,
		Entries: []DriftEntry{
			{Field: "image", Kind: DriftKindImage, Declared: "v1", Observed: "v2"},
		},
	}
}

func TestNewEscalator_NotNil(t *testing.T) {
	e := NewEscalator(DefaultEscalationPolicy())
	if e == nil {
		t.Fatal("expected non-nil escalator")
	}
}

func TestNewEscalator_ZeroPolicy_UsesDefaults(t *testing.T) {
	e := NewEscalator(EscalationPolicy{})
	if e.policy.WarnAfter != DefaultEscalationPolicy().WarnAfter {
		t.Errorf("expected default WarnAfter, got %v", e.policy.WarnAfter)
	}
}

func TestEscalator_Evaluate_CleanResult_None(t *testing.T) {
	e := NewEscalator(DefaultEscalationPolicy())
	result := DriftResult{Service: "svc-a"}
	if level := e.Evaluate(result); level != EscalationNone {
		t.Errorf("expected none, got %s", level)
	}
}

func TestEscalator_Evaluate_FirstDrift_None(t *testing.T) {
	e := NewEscalator(DefaultEscalationPolicy())
	result := escalatedResult("svc-a")
	if level := e.Evaluate(result); level != EscalationNone {
		t.Errorf("expected none on first observation, got %s", level)
	}
}

func TestEscalator_Evaluate_WarnAfterThreshold(t *testing.T) {
	policy := EscalationPolicy{
		WarnAfter:     10 * time.Millisecond,
		CriticalAfter: 1 * time.Hour,
	}
	e := NewEscalator(policy)
	result := escalatedResult("svc-b")
	e.Evaluate(result) // seed
	time.Sleep(20 * time.Millisecond)
	if level := e.Evaluate(result); level != EscalationWarning {
		t.Errorf("expected warning, got %s", level)
	}
}

func TestEscalator_Evaluate_CriticalAfterThreshold(t *testing.T) {
	policy := EscalationPolicy{
		WarnAfter:     5 * time.Millisecond,
		CriticalAfter: 10 * time.Millisecond,
	}
	e := NewEscalator(policy)
	result := escalatedResult("svc-c")
	e.Evaluate(result)
	time.Sleep(20 * time.Millisecond)
	if level := e.Evaluate(result); level != EscalationCritical {
		t.Errorf("expected critical, got %s", level)
	}
}

func TestEscalator_CleanResult_ClearsState(t *testing.T) {
	policy := EscalationPolicy{
		WarnAfter:     5 * time.Millisecond,
		CriticalAfter: 1 * time.Hour,
	}
	e := NewEscalator(policy)
	drifted := escalatedResult("svc-d")
	clean := DriftResult{Service: "svc-d"}
	e.Evaluate(drifted)
	time.Sleep(10 * time.Millisecond)
	e.Evaluate(clean)
	// After clean, first drift observation should reset
	if level := e.Evaluate(drifted); level != EscalationNone {
		t.Errorf("expected none after clean reset, got %s", level)
	}
}

func TestEscalator_Reset_ClearsEntry(t *testing.T) {
	policy := EscalationPolicy{
		WarnAfter:     5 * time.Millisecond,
		CriticalAfter: 1 * time.Hour,
	}
	e := NewEscalator(policy)
	result := escalatedResult("svc-e")
	e.Evaluate(result)
	time.Sleep(10 * time.Millisecond)
	e.Reset("svc-e")
	if level := e.Evaluate(result); level != EscalationNone {
		t.Errorf("expected none after reset, got %s", level)
	}
}

func TestEscalationLevel_String(t *testing.T) {
	cases := map[EscalationLevel]string{
		EscalationNone:     "none",
		EscalationWarning:  "warning",
		EscalationCritical: "critical",
	}
	for level, want := range cases {
		if got := level.String(); got != want {
			t.Errorf("EscalationLevel(%d).String() = %q, want %q", level, got, want)
		}
	}
}

func TestEscalator_Summary_ContainsCount(t *testing.T) {
	e := NewEscalator(DefaultEscalationPolicy())
	e.Evaluate(escalatedResult("svc-x"))
	s := e.Summary()
	if s == "" {
		t.Error("expected non-empty summary")
	}
}
