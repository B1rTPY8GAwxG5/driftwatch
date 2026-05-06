package drift

import (
	"bytes"
	"strings"
	"testing"
)

func TestRollupFormatter_EmptyRollup(t *testing.T) {
	f := NewRollupFormatter()
	var buf bytes.Buffer
	r := BuildRollup(nil)
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Drift Rollup") {
		t.Errorf("expected header in output, got: %s", out)
	}
}

func TestRollupFormatter_CleanService(t *testing.T) {
	f := NewRollupFormatter()
	var buf bytes.Buffer
	results := []DriftResult{makeRollupResult("api", nil)}
	r := BuildRollup(results)
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "api") {
		t.Errorf("expected service name in output")
	}
	if !strings.Contains(out, "clean") {
		t.Errorf("expected clean status in output")
	}
}

func TestRollupFormatter_DriftedService(t *testing.T) {
	f := NewRollupFormatter()
	var buf bytes.Buffer
	results := []DriftResult{
		makeRollupResult("worker", []DriftEntry{{Kind: KindImage, Field: "image"}}),
	}
	r := BuildRollup(results)
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "drifted") {
		t.Errorf("expected drifted status in output, got: %s", out)
	}
	if !strings.Contains(out, "image") {
		t.Errorf("expected image kind in output, got: %s", out)
	}
}

func TestRollupFormatter_MultipleServices(t *testing.T) {
	f := NewRollupFormatter()
	var buf bytes.Buffer
	results := []DriftResult{
		makeRollupResult("svc-a", nil),
		makeRollupResult("svc-b", []DriftEntry{{Kind: KindReplicas}}),
		makeRollupResult("svc-c", nil),
	}
	r := BuildRollup(results)
	if err := f.Format(&buf, r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, name := range []string{"svc-a", "svc-b", "svc-c"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %s in output", name)
		}
	}
}
