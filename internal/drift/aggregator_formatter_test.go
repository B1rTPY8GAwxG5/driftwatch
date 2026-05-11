package drift

import (
	"strings"
	"testing"
	"bytes"
)

func TestAggregatorFormatter_NoDrift(t *testing.T) {
	f := NewAggregatorFormatter()
	a := NewAggregator(AggregatorModeAny)
	res := a.Aggregate([]DriftResult{
		{Service: "svc-a", Entries: nil},
	})
	var buf bytes.Buffer
	if err := f.Format(&buf, res); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "clean") {
		t.Errorf("expected 'clean' in output, got: %s", out)
	}
	if !strings.Contains(out, "any") {
		t.Errorf("expected mode 'any' in output, got: %s", out)
	}
}

func TestAggregatorFormatter_WithDrift(t *testing.T) {
	f := NewAggregatorFormatter()
	a := NewAggregator(AggregatorModeAny)
	res := a.Aggregate([]DriftResult{
		{Service: "svc-b", Entries: []DriftEntry{{Kind: DriftKindImage, Field: "image", Declared: "v1", Observed: "v2"}}},
	})
	var buf bytes.Buffer
	if err := f.Format(&buf, res); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "DRIFTED") {
		t.Errorf("expected 'DRIFTED' in output, got: %s", out)
	}
	if !strings.Contains(out, "svc-b") {
		t.Errorf("expected service name in output, got: %s", out)
	}
}

func TestAggregatorFormatter_AllMode(t *testing.T) {
	f := NewAggregatorFormatter()
	a := NewAggregator(AggregatorModeAll)
	res := a.Aggregate([]DriftResult{
		{Service: "svc-a", Entries: nil},
		{Service: "svc-b", Entries: []DriftEntry{{Kind: DriftKindImage, Field: "image", Declared: "v1", Observed: "v2"}}},
	})
	var buf bytes.Buffer
	if err := f.Format(&buf, res); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "all") {
		t.Errorf("expected mode 'all' in output, got: %s", out)
	}
	// partial drift in all-mode => clean
	if !strings.Contains(out, "clean") {
		t.Errorf("expected 'clean' for partial drift in all-mode, got: %s", out)
	}
}

func TestAggregatorFormatter_TotalServicesCount(t *testing.T) {
	f := NewAggregatorFormatter()
	a := NewAggregator(AggregatorModeAny)
	res := a.Aggregate([]DriftResult{
		{Service: "svc-a"},
		{Service: "svc-b"},
		{Service: "svc-c"},
	})
	var buf bytes.Buffer
	if err := f.Format(&buf, res); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "3") {
		t.Errorf("expected total count 3 in output, got: %s", buf.String())
	}
}
