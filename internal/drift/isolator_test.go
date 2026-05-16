package drift

import (
	"testing"
	"time"
)

func TestDefaultIsolationPolicy_Values(t *testing.T) {
	p := DefaultIsolationPolicy()
	if p.Threshold != 3 {
		t.Errorf("expected threshold 3, got %d", p.Threshold)
	}
	if p.Duration != 10*time.Minute {
		t.Errorf("expected duration 10m, got %v", p.Duration)
	}
}

func TestNewIsolator_ZeroPolicy_UsesDefaults(t *testing.T) {
	iso := NewIsolator(IsolationPolicy{})
	if iso.policy.Threshold != 3 {
		t.Errorf("expected default threshold 3, got %d", iso.policy.Threshold)
	}
}

func TestIsolator_NotIsolated_Initially(t *testing.T) {
	iso := NewIsolator(DefaultIsolationPolicy())
	if iso.IsIsolated("svc-a") {
		t.Error("expected service to not be isolated initially")
	}
}

func TestIsolator_BelowThreshold_NotIsolated(t *testing.T) {
	iso := NewIsolator(IsolationPolicy{Threshold: 3, Duration: time.Minute})
	r := DriftResult{Service: "svc-a", Entries: []DriftEntry{{Kind: KindImage}}}
	iso.Record(r)
	iso.Record(r)
	if iso.IsIsolated("svc-a") {
		t.Error("expected service not isolated below threshold")
	}
}

func TestIsolator_AtThreshold_Isolated(t *testing.T) {
	iso := NewIsolator(IsolationPolicy{Threshold: 3, Duration: time.Minute})
	r := DriftResult{Service: "svc-a", Entries: []DriftEntry{{Kind: KindImage}}}
	iso.Record(r)
	iso.Record(r)
	iso.Record(r)
	if !iso.IsIsolated("svc-a") {
		t.Error("expected service to be isolated at threshold")
	}
}

func TestIsolator_CleanResult_ResetsIsolation(t *testing.T) {
	iso := NewIsolator(IsolationPolicy{Threshold: 2, Duration: time.Minute})
	drifted := DriftResult{Service: "svc-b", Entries: []DriftEntry{{Kind: KindImage}}}
	clean := DriftResult{Service: "svc-b"}
	iso.Record(drifted)
	iso.Record(drifted)
	iso.Record(clean)
	if iso.IsIsolated("svc-b") {
		t.Error("expected isolation to be cleared after clean result")
	}
}

func TestIsolator_IsolationExpires(t *testing.T) {
	iso := NewIsolator(IsolationPolicy{Threshold: 1, Duration: time.Millisecond})
	r := DriftResult{Service: "svc-c", Entries: []DriftEntry{{Kind: KindReplicas}}}
	iso.Record(r)
	if !iso.IsIsolated("svc-c") {
		t.Error("expected service to be isolated")
	}
	time.Sleep(5 * time.Millisecond)
	if iso.IsIsolated("svc-c") {
		t.Error("expected isolation to have expired")
	}
}

func TestIsolator_Reset_ClearsState(t *testing.T) {
	iso := NewIsolator(IsolationPolicy{Threshold: 1, Duration: time.Minute})
	r := DriftResult{Service: "svc-d", Entries: []DriftEntry{{Kind: KindImage}}}
	iso.Record(r)
	iso.Reset("svc-d")
	if iso.IsIsolated("svc-d") {
		t.Error("expected state to be cleared after reset")
	}
}
