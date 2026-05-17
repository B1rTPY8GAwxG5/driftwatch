package drift

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewTracer_NotNil(t *testing.T) {
	tr := NewTracer()
	if tr == nil {
		t.Fatal("expected non-nil Tracer")
	}
}

func TestTracer_Len_InitiallyZero(t *testing.T) {
	tr := NewTracer()
	if tr.Len() != 0 {
		t.Fatalf("expected 0, got %d", tr.Len())
	}
}

func TestTracer_Record_IncrementsLen(t *testing.T) {
	tr := NewTracer()
	tr.Record("detect", "svc-a", "ok", time.Millisecond, nil)
	if tr.Len() != 1 {
		t.Fatalf("expected 1, got %d", tr.Len())
	}
}

func TestTracer_Events_ReturnsCopy(t *testing.T) {
	tr := NewTracer()
	tr.Record("detect", "svc-a", "ok", time.Millisecond, nil)
	events := tr.Events()
	events[0].Service = "mutated"
	if tr.Events()[0].Service == "mutated" {
		t.Fatal("Events should return a copy, not a reference")
	}
}

func TestTracer_Record_WithError(t *testing.T) {
	tr := NewTracer()
	tr.Record("compare", "svc-b", "", 0, errors.New("timeout"))
	evts := tr.Events()
	if evts[0].Err == nil {
		t.Fatal("expected error to be recorded")
	}
}

func TestTracer_Reset_ClearsEvents(t *testing.T) {
	tr := NewTracer()
	tr.Record("detect", "svc-a", "ok", time.Millisecond, nil)
	tr.Reset()
	if tr.Len() != 0 {
		t.Fatalf("expected 0 after reset, got %d", tr.Len())
	}
}

func TestTracer_WriteTo_ContainsStage(t *testing.T) {
	tr := NewTracer()
	tr.Record("pipeline", "svc-x", "processed", 2*time.Millisecond, nil)
	var buf bytes.Buffer
	tr.WriteTo(&buf)
	if !strings.Contains(buf.String(), "pipeline") {
		t.Errorf("expected output to contain stage name, got: %s", buf.String())
	}
}

func TestTracer_WriteTo_ContainsErrorMarker(t *testing.T) {
	tr := NewTracer()
	tr.Record("load", "svc-y", "", 0, errors.New("not found"))
	var buf bytes.Buffer
	tr.WriteTo(&buf)
	if !strings.Contains(buf.String(), "ERR") {
		t.Errorf("expected ERR marker in output, got: %s", buf.String())
	}
}

func TestTraceEvent_String_NoError(t *testing.T) {
	e := TraceEvent{
		Stage:     "score",
		Service:   "svc-z",
		Message:   "done",
		Timestamp: time.Now().UTC(),
		Duration:  5 * time.Millisecond,
	}
	s := e.String()
	if !strings.Contains(s, "score") || !strings.Contains(s, "svc-z") {
		t.Errorf("unexpected string output: %s", s)
	}
}
