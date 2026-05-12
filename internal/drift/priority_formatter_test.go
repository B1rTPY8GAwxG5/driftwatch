package drift

import (
	"strings"
	"testing"
)

func TestPriorityFormatter_EmptyResults(t *testing.T) {
	f := NewPriorityFormatter(nil)
	var sb strings.Builder
	if err := f.Format(&sb, []DriftResult{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sb.String(), "no results") {
		t.Errorf("expected 'no results' message, got: %s", sb.String())
	}
}

func TestPriorityFormatter_CleanResult_LowLabel(t *testing.T) {
	f := NewPriorityFormatter(nil)
	var sb strings.Builder
	results := []DriftResult{{Service: "svc", Entries: []DriftEntry{}}}
	if err := f.Format(&sb, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "[LOW]") {
		t.Errorf("expected [LOW] label for clean result, got: %s", out)
	}
	if !strings.Contains(out, "clean") {
		t.Errorf("expected 'clean' status, got: %s", out)
	}
}

func TestPriorityFormatter_DriftedResult_HighLabel(t *testing.T) {
	f := NewPriorityFormatter(nil)
	var sb strings.Builder
	results := []DriftResult{
		{
			Service: "api",
			Entries: []DriftEntry{
				{Kind: DriftKindImage, Field: "image", Declared: "v1", Observed: "v2"},
			},
		},
	}
	if err := f.Format(&sb, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "[HIGH]") {
		t.Errorf("expected [HIGH] label, got: %s", out)
	}
	if !strings.Contains(out, "drifted") {
		t.Errorf("expected 'drifted' status, got: %s", out)
	}
}

func TestPriorityFormatter_EntryDetails_Rendered(t *testing.T) {
	f := NewPriorityFormatter(nil)
	var sb strings.Builder
	results := []DriftResult{
		{
			Service: "api",
			Entries: []DriftEntry{
				{Kind: DriftKindEnv, Field: "LOG_LEVEL", Declared: "info", Observed: "debug"},
			},
		},
	}
	if err := f.Format(&sb, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "LOG_LEVEL") {
		t.Errorf("expected field name in output, got: %s", out)
	}
	if !strings.Contains(out, "info") || !strings.Contains(out, "debug") {
		t.Errorf("expected declared/observed values in output, got: %s", out)
	}
}

func TestPriorityFormatter_NilPrioritizer_UsesDefault(t *testing.T) {
	f := NewPriorityFormatter(nil)
	if f == nil {
		t.Fatal("expected non-nil formatter")
	}
	if f.prioritizer == nil {
		t.Fatal("expected default prioritizer to be set")
	}
}
