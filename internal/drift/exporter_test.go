package drift

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

var exportedResult = DriftResult{
	Service: "api-gateway",
	Entries: []DriftEntry{
		{Kind: DriftKindImage, Field: "image", Declared: "nginx:1.24", Actual: "nginx:1.21"},
	},
}

var cleanExportResult = DriftResult{
	Service:  "clean-service",
	Entries:  []DriftEntry{},
}

func TestNewExporter_JSON(t *testing.T) {
	e, err := NewExporter(ExportFormatJSON, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e == nil {
		t.Fatal("expected non-nil exporter")
	}
}

func TestNewExporter_Text(t *testing.T) {
	e, err := NewExporter(ExportFormatText, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e == nil {
		t.Fatal("expected non-nil exporter")
	}
}

func TestNewExporter_Unknown(t *testing.T) {
	_, err := NewExporter(ExportFormat("xml"), &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestNewExporter_NilWriter(t *testing.T) {
	_, err := NewExporter(ExportFormatJSON, nil)
	if err == nil {
		t.Fatal("expected error for nil writer")
	}
}

func TestExporter_Export_JSON_HasDrift(t *testing.T) {
	var buf bytes.Buffer
	e, _ := NewExporter(ExportFormatJSON, &buf)
	if err := e.Export(exportedResult); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var rec ExportRecord
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if rec.Service != "api-gateway" {
		t.Errorf("expected service api-gateway, got %s", rec.Service)
	}
	if !rec.HasDrift {
		t.Error("expected has_drift=true")
	}
	if len(rec.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(rec.Entries))
	}
}

func TestExporter_Export_JSON_Clean(t *testing.T) {
	var buf bytes.Buffer
	e, _ := NewExporter(ExportFormatJSON, &buf)
	if err := e.Export(cleanExportResult); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var rec ExportRecord
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if rec.HasDrift {
		t.Error("expected has_drift=false")
	}
}

func TestExporter_Export_Text_ContainsService(t *testing.T) {
	var buf bytes.Buffer
	e, _ := NewExporter(ExportFormatText, &buf)
	if err := e.Export(exportedResult); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "api-gateway") {
		t.Errorf("expected service name in output, got: %s", out)
	}
	if !strings.Contains(out, "drifted") {
		t.Errorf("expected status=drifted in output, got: %s", out)
	}
}

func TestExporter_Export_Text_CleanStatus(t *testing.T) {
	var buf bytes.Buffer
	e, _ := NewExporter(ExportFormatText, &buf)
	if err := e.Export(cleanExportResult); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "clean") {
		t.Errorf("expected status=clean in output, got: %s", out)
	}
}

func TestExporter_Export_JSON_EntryFields(t *testing.T) {
	var buf bytes.Buffer
	e, _ := NewExporter(ExportFormatJSON, &buf)
	if err := e.Export(exportedResult); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var rec ExportRecord
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	entry := rec.Entries[0]
	if entry.Field != "image" {
		t.Errorf("expected field=image, got %s", entry.Field)
	}
	if entry.Declared != "nginx:1.24" {
		t.Errorf("expected declared=nginx:1.24, got %s", entry.Declared)
	}
	if entry.Actual != "nginx:1.21" {
		t.Errorf("expected actual=nginx:1.21, got %s", entry.Actual)
	}
}
