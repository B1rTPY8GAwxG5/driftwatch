package drift

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func heatmapDriftedResult(service string, at time.Time) DriftResult {
	return DriftResult{
		Service:   service,
		CheckedAt: at,
		Entries: []DriftEntry{
			{Kind: DriftKindImage, Field: "image", Declared: "a", Observed: "b"},
		},
	}
}

func heatmapCleanResult(service string, at time.Time) DriftResult {
	return DriftResult{Service: service, CheckedAt: at}
}

func TestNewDriftHeatmap_NotNil(t *testing.T) {
	h := NewDriftHeatmap()
	if h == nil {
		t.Fatal("expected non-nil heatmap")
	}
}

func TestDriftHeatmap_Record_CleanIgnored(t *testing.T) {
	h := NewDriftHeatmap()
	h.Record(heatmapCleanResult("svc-a", time.Now()))
	if len(h.Cells()) != 0 {
		t.Errorf("expected 0 cells, got %d", len(h.Cells()))
	}
}

func TestDriftHeatmap_Record_DriftedCounted(t *testing.T) {
	h := NewDriftHeatmap()
	now := time.Now().UTC()
	h.Record(heatmapDriftedResult("svc-a", now))
	cells := h.Cells()
	if len(cells) != 1 {
		t.Fatalf("expected 1 cell, got %d", len(cells))
	}
	if cells[0].Count != 1 {
		t.Errorf("expected count 1, got %d", cells[0].Count)
	}
}

func TestDriftHeatmap_Record_AccumulatesWithinHour(t *testing.T) {
	h := NewDriftHeatmap()
	now := time.Now().UTC()
	h.Record(heatmapDriftedResult("svc-a", now))
	h.Record(heatmapDriftedResult("svc-a", now.Add(10*time.Minute)))
	cells := h.Cells()
	if len(cells) != 1 {
		t.Fatalf("expected 1 cell (same hour bucket), got %d", len(cells))
	}
	if cells[0].Count != 2 {
		t.Errorf("expected count 2, got %d", cells[0].Count)
	}
}

func TestDriftHeatmap_Record_SeparateBuckets(t *testing.T) {
	h := NewDriftHeatmap()
	now := time.Now().UTC().Truncate(time.Hour)
	h.Record(heatmapDriftedResult("svc-a", now))
	h.Record(heatmapDriftedResult("svc-a", now.Add(2*time.Hour)))
	if len(h.Cells()) != 2 {
		t.Errorf("expected 2 cells, got %d", len(h.Cells()))
	}
}

func TestDriftHeatmap_WriteTo_NoDrift(t *testing.T) {
	h := NewDriftHeatmap()
	var buf bytes.Buffer
	if err := h.WriteTo(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no drift") {
		t.Errorf("expected 'no drift' message, got: %s", buf.String())
	}
}

func TestDriftHeatmap_WriteTo_WithDrift(t *testing.T) {
	h := NewDriftHeatmap()
	now := time.Now().UTC()
	h.Record(heatmapDriftedResult("svc-b", now))
	var buf bytes.Buffer
	if err := h.WriteTo(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "svc-b") {
		t.Errorf("expected service name in output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "SERVICE") {
		t.Errorf("expected header in output, got: %s", buf.String())
	}
}
