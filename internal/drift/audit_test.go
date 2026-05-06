package drift

import (
	"bytes"
	"strings"
	"testing"
)

var cleanAuditResult = DriftResult{
	Service: "api-server",
	Entries: []DriftEntry{},
}

var driftedAuditResult = DriftResult{
	Service: "api-server",
	Entries: []DriftEntry{
		{Kind: KindImage, Field: "image", Declared: "nginx:1.24", Observed: "nginx:1.25"},
	},
}

func TestNewAuditLog_Empty(t *testing.T) {
	log := NewAuditLog()
	if log == nil {
		t.Fatal("expected non-nil AuditLog")
	}
	if log.Len() != 0 {
		t.Fatalf("expected 0 events, got %d", log.Len())
	}
}

func TestAuditLog_Record_CleanResult(t *testing.T) {
	log := NewAuditLog()
	log.Record(cleanAuditResult)

	if log.Len() != 1 {
		t.Fatalf("expected 1 event, got %d", log.Len())
	}
	events := log.Events()
	if events[0].Drifted {
		t.Error("expected Drifted=false for clean result")
	}
	if events[0].Message != "no drift detected" {
		t.Errorf("unexpected message: %s", events[0].Message)
	}
}

func TestAuditLog_Record_DriftedResult(t *testing.T) {
	log := NewAuditLog()
	log.Record(driftedAuditResult)

	events := log.Events()
	if !events[0].Drifted {
		t.Error("expected Drifted=true for drifted result")
	}
	if !strings.Contains(events[0].Message, "1 field") {
		t.Errorf("expected message to mention field count, got: %s", events[0].Message)
	}
}

func TestAuditLog_Events_ReturnsCopy(t *testing.T) {
	log := NewAuditLog()
	log.Record(cleanAuditResult)

	e1 := log.Events()
	e1[0].Service = "mutated"
	e2 := log.Events()

	if e2[0].Service == "mutated" {
		t.Error("Events() should return a copy, not a reference")
	}
}

func TestAuditLog_WriteTo_NoDrift(t *testing.T) {
	log := NewAuditLog()
	log.Record(cleanAuditResult)

	var buf bytes.Buffer
	if err := log.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "api-server") {
		t.Errorf("expected service name in output, got: %s", out)
	}
	if !strings.Contains(out, "drifted=false") {
		t.Errorf("expected drifted=false in output, got: %s", out)
	}
}

func TestAuditLog_WriteTo_WithDrift(t *testing.T) {
	log := NewAuditLog()
	log.Record(driftedAuditResult)

	var buf bytes.Buffer
	if err := log.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "drifted=true") {
		t.Errorf("expected drifted=true in output, got: %s", out)
	}
}
