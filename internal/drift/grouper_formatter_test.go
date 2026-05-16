package drift

import (
	"strings"
	"testing"
)

func TestGrouperFormatter_EmptyResults(t *testing.T) {
	f := NewGrouperFormatter(nil)
	var sb strings.Builder
	if err := f.Format(&sb, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sb.String(), "no results") {
		t.Errorf("expected 'no results' message, got %q", sb.String())
	}
}

func TestGrouperFormatter_NilGrouper_UsesDefault(t *testing.T) {
	f := NewGrouperFormatter(nil)
	if f.grouper == nil {
		t.Fatal("expected non-nil grouper")
	}
	if f.grouper.Mode() != GroupByService {
		t.Errorf("expected GroupByService default, got %q", f.grouper.Mode())
	}
}

func TestGrouperFormatter_CleanResult(t *testing.T) {
	f := NewGrouperFormatter(NewGrouper(GroupByService))
	results := []DriftResult{{Service: "api", Entries: nil}}
	var sb strings.Builder
	if err := f.Format(&sb, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "API") {
		t.Errorf("expected group header with API, got %q", out)
	}
	if !strings.Contains(out, "clean") {
		t.Errorf("expected 'clean' for drift-free result, got %q", out)
	}
}

func TestGrouperFormatter_DriftedResult(t *testing.T) {
	f := NewGrouperFormatter(NewGrouper(GroupByService))
	results := []DriftResult{
		{Service: "web", Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "nginx:1.24", Actual: "nginx:1.25"},
		}},
	}
	var sb strings.Builder
	if err := f.Format(&sb, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "nginx:1.24") {
		t.Errorf("expected declared value in output, got %q", out)
	}
	if !strings.Contains(out, "nginx:1.25") {
		t.Errorf("expected actual value in output, got %q", out)
	}
}

func TestGrouperFormatter_MultipleGroups_Sorted(t *testing.T) {
	f := NewGrouperFormatter(NewGrouper(GroupByService))
	results := []DriftResult{
		{Service: "zoo"}, {Service: "aardvark"}, {Service: "monkey"},
	}
	var sb strings.Builder
	if err := f.Format(&sb, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	aIdx := strings.Index(out, "AARDVARK")
	mIdx := strings.Index(out, "MONKEY")
	zIdx := strings.Index(out, "ZOO")
	if !(aIdx < mIdx && mIdx < zIdx) {
		t.Errorf("groups not in sorted order in output")
	}
}
