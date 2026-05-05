package drift_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/driftwatch/internal/drift"
)

func TestNewFormatter_Text(t *testing.T) {
	f, err := drift.NewFormatter(drift.FormatText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil formatter")
	}
}

func TestNewFormatter_Unknown(t *testing.T) {
	_, err := drift.NewFormatter("xml")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestTextFormatter_NoDrift(t *testing.T) {
	report := &drift.DriftReport{
		Service: "api",
		Entries: []drift.DriftEntry{},
	}

	var buf bytes.Buffer
	f := drift.NewTextFormatter()
	if err := f.Format(&buf, report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "No drift detected") {
		t.Errorf("expected no-drift message, got: %q", buf.String())
	}
}

func TestTextFormatter_WithDrift(t *testing.T) {
	report := &drift.DriftReport{
		Service: "api",
		Entries: []drift.DriftEntry{
			{Kind: drift.KindImage, Field: "image", Expected: "nginx:1.25", Actual: "nginx:1.24"},
			{Kind: drift.KindReplicas, Field: "replicas", Expected: 3, Actual: 1},
		},
	}

	var buf bytes.Buffer
	f := drift.NewTextFormatter()
	if err := f.Format(&buf, report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"FIELD", "EXPECTED", "ACTUAL", "image", "replicas"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got:\n%s", want, out)
		}
	}
}

func TestTextFormatter_DefaultFormat(t *testing.T) {
	f, err := drift.NewFormatter("")
	if err != nil {
		t.Fatalf("unexpected error for empty format: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil formatter for default")
	}
}
