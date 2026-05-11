package drift

import (
	"strings"
	"testing"
)

func cleanAggResult() DriftResult {
	return DriftResult{Service: "svc-a", Entries: nil}
}

func driftedAggResult(name string) DriftResult {
	return DriftResult{
		Service: name,
		Entries: []DriftEntry{{Kind: DriftKindImage, Field: "image", Declared: "v1", Observed: "v2"}},
	}
}

func TestNewAggregator_DefaultMode(t *testing.T) {
	a := NewAggregator("")
	if a.mode != AggregatorModeAny {
		t.Fatalf("expected any, got %s", a.mode)
	}
}

func TestAggregator_Aggregate_NoDrift(t *testing.T) {
	a := NewAggregator(AggregatorModeAny)
	res := a.Aggregate([]DriftResult{cleanAggResult(), cleanAggResult()})
	if res.HasDrift() {
		t.Fatal("expected no drift")
	}
	if len(res.Services) != 0 {
		t.Fatalf("expected empty services, got %v", res.Services)
	}
}

func TestAggregator_Aggregate_AnyMode_OneDrifted(t *testing.T) {
	a := NewAggregator(AggregatorModeAny)
	res := a.Aggregate([]DriftResult{cleanAggResult(), driftedAggResult("svc-b")})
	if !res.HasDrift() {
		t.Fatal("expected drift")
	}
	if len(res.Services) != 1 || res.Services[0] != "svc-b" {
		t.Fatalf("unexpected services: %v", res.Services)
	}
}

func TestAggregator_Aggregate_AllMode_PartialDrift(t *testing.T) {
	a := NewAggregator(AggregatorModeAll)
	res := a.Aggregate([]DriftResult{cleanAggResult(), driftedAggResult("svc-b")})
	if res.HasDrift() {
		t.Fatal("expected no drift in all-mode with only partial drift")
	}
}

func TestAggregator_Aggregate_AllMode_AllDrifted(t *testing.T) {
	a := NewAggregator(AggregatorModeAll)
	res := a.Aggregate([]DriftResult{driftedAggResult("svc-a"), driftedAggResult("svc-b")})
	if !res.HasDrift() {
		t.Fatal("expected drift")
	}
}

func TestAggregatedResult_Summary_NoDrift(t *testing.T) {
	a := NewAggregator(AggregatorModeAny)
	res := a.Aggregate([]DriftResult{cleanAggResult()})
	s := res.Summary()
	if !strings.Contains(s, "clean") {
		t.Fatalf("expected 'clean' in summary, got: %s", s)
	}
}

func TestAggregatedResult_Summary_WithDrift(t *testing.T) {
	a := NewAggregator(AggregatorModeAny)
	res := a.Aggregate([]DriftResult{driftedAggResult("svc-x")})
	s := res.Summary()
	if !strings.Contains(s, "svc-x") {
		t.Fatalf("expected service name in summary, got: %s", s)
	}
}
