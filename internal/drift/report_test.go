package drift

import (
	"bytes"
	"strings"
	"testing"
)

func TestReport_HasDrift_False(t *testing.T) {
	r := &Report{Service: "svc", Results: nil}
	if r.HasDrift() {
		t.Error("expected no drift")
	}
}

func TestReport_HasDrift_True(t *testing.T) {
	r := &Report{
		Service: "svc",
		Results: []DriftResult{{Field: "image", Status: StatusDrifted}},
	}
	if !r.HasDrift() {
		t.Error("expected drift")
	}
}

func TestReport_Summary_NoDrift(t *testing.T) {
	r := &Report{Service: "api", Results: nil}
	if !strings.Contains(r.Summary(), "[OK]") {
		t.Errorf("unexpected summary: %s", r.Summary())
	}
}

func TestReport_Summary_WithDrift(t *testing.T) {
	r := &Report{
		Service: "api",
		Results: []DriftResult{
			{Field: "image", Status: StatusDrifted, Declared: "v1", Live: "v2"},
		},
	}
	summary := r.Summary()
	if !strings.Contains(summary, "[DRIFT]") {
		t.Errorf("expected DRIFT in summary, got: %s", summary)
	}
	if !strings.Contains(summary, "1 field") {
		t.Errorf("expected field count in summary, got: %s", summary)
	}
}

func TestReport_Summary_MultipleFields(t *testing.T) {
	r := &Report{
		Service: "api",
		Results: []DriftResult{
			{Field: "image", Status: StatusDrifted, Declared: "v1", Live: "v2"},
			{Field: "replicas", Status: StatusDrifted, Declared: "2", Live: "3"},
		},
	}
	summary := r.Summary()
	if !strings.Contains(summary, "2 field") {
		t.Errorf("expected '2 field' in summary, got: %s", summary)
	}
}

func TestReport_WriteTo_NoDrift(t *testing.T) {
	r := &Report{Service: "api", Results: nil}
	var buf bytes.Buffer
	if err := r.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if !strings.Contains(buf.String(), "[OK]") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestBuildReport_Integration(t *testing.T) {
	declared := baseService()
	live := baseService()
	live.Image = "myapp:old"

	report := BuildReport(declared, live)
	if report.Service != "api" {
		t.Errorf("unexpected service name: %s", report.Service)
	}
	if !report.HasDrift() {
		t.Error("expected drift in integration test")
	}

	var buf bytes.Buffer
	if err := report.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if !strings.Contains(buf.String(), "image") {
		t.Errorf("expected 'image' in output: %s", buf.String())
	}
}
