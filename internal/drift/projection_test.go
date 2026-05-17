package drift

import (
	"strings"
	"testing"
)

func projResult(service string, entries ...DriftEntry) DriftResult {
	return DriftResult{Service: service, Entries: entries}
}

func TestNewProjection_DefaultFields(t *testing.T) {
	p := NewProjection()
	if len(p.Headers()) != 7 {
		t.Fatalf("expected 7 headers, got %d", len(p.Headers()))
	}
}

func TestNewProjection_CustomFields(t *testing.T) {
	p := NewProjection(ProjectionFieldService, ProjectionFieldDrifted)
	if len(p.Headers()) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(p.Headers()))
	}
}

func TestProjection_Project_NoDrift_SingleRow(t *testing.T) {
	p := NewProjection()
	r := projResult("svc-a")
	rows := p.Project(r)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0][string(ProjectionFieldService)] != "svc-a" {
		t.Errorf("expected service svc-a")
	}
	if rows[0][string(ProjectionFieldDrifted)] != "false" {
		t.Errorf("expected drifted=false")
	}
}

func TestProjection_Project_WithEntries_MultipleRows(t *testing.T) {
	p := NewProjection()
	e1 := DriftEntry{Kind: KindImage, Field: "image", Expected: "v1", Actual: "v2"}
	e2 := DriftEntry{Kind: KindReplicas, Field: "replicas", Expected: 1, Actual: 3}
	r := projResult("svc-b", e1, e2)
	rows := p.Project(r)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	for _, row := range rows {
		if row[string(ProjectionFieldDrifted)] != "true" {
			t.Errorf("expected drifted=true")
		}
		if row[string(ProjectionFieldService)] != "svc-b" {
			t.Errorf("expected service svc-b")
		}
	}
}

func TestProjection_Project_FieldValues(t *testing.T) {
	p := NewProjection(ProjectionFieldKind, ProjectionFieldField, ProjectionFieldExpected, ProjectionFieldActual)
	e := DriftEntry{Kind: KindImage, Field: "image", Expected: "nginx:1", Actual: "nginx:2"}
	rows := p.Project(projResult("svc", e))
	if rows[0][string(ProjectionFieldKind)] != string(KindImage) {
		t.Errorf("kind mismatch")
	}
	if rows[0][string(ProjectionFieldExpected)] != "nginx:1" {
		t.Errorf("expected mismatch")
	}
	if rows[0][string(ProjectionFieldActual)] != "nginx:2" {
		t.Errorf("actual mismatch")
	}
}

func TestProjection_ProjectAll_CombinesResults(t *testing.T) {
	p := NewProjection()
	r1 := projResult("svc-x")
	e := DriftEntry{Kind: KindImage, Field: "image", Expected: "a", Actual: "b"}
	r2 := projResult("svc-y", e)
	rows := p.ProjectAll([]DriftResult{r1, r2})
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestProjection_Headers_Order(t *testing.T) {
	p := NewProjection(ProjectionFieldService, ProjectionFieldKind)
	h := p.Headers()
	if h[0] != "service" || h[1] != "kind" {
		t.Errorf("unexpected header order: %v", h)
	}
}

func TestFormatTable_ContainsHeaders(t *testing.T) {
	p := NewProjection(ProjectionFieldService, ProjectionFieldDrifted)
	rows := p.Project(projResult("svc-z"))
	out := FormatTable(p.Headers(), rows)
	if !strings.Contains(out, "service") {
		t.Errorf("expected 'service' header in output")
	}
	if !strings.Contains(out, "svc-z") {
		t.Errorf("expected 'svc-z' in output")
	}
}

func TestFormatTable_SortedByService(t *testing.T) {
	p := NewProjection(ProjectionFieldService)
	rows := p.ProjectAll([]DriftResult{
		projResult("zzz"),
		projResult("aaa"),
	})
	out := FormatTable(p.Headers(), rows)
	idxA := strings.Index(out, "aaa")
	idxZ := strings.Index(out, "zzz")
	if idxA > idxZ {
		t.Errorf("expected aaa before zzz in sorted table")
	}
}
