package drift

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestDefaultFlushPolicy_Values(t *testing.T) {
	p := DefaultFlushPolicy()
	if p.MaxSize != 50 {
		t.Errorf("expected MaxSize 50, got %d", p.MaxSize)
	}
	if p.MaxAge != 5*time.Minute {
		t.Errorf("expected MaxAge 5m, got %v", p.MaxAge)
	}
}

func TestNewFlusher_NotNil(t *testing.T) {
	f := NewFlusher(DefaultFlushPolicy(), &bytes.Buffer{})
	if f == nil {
		t.Fatal("expected non-nil Flusher")
	}
}

func TestNewFlusher_ZeroPolicy_UsesDefaults(t *testing.T) {
	f := NewFlusher(FlushPolicy{}, &bytes.Buffer{})
	if f.policy.MaxSize != 50 {
		t.Errorf("expected default MaxSize 50, got %d", f.policy.MaxSize)
	}
}

func TestFlusher_Len_InitiallyZero(t *testing.T) {
	f := NewFlusher(DefaultFlushPolicy(), &bytes.Buffer{})
	if f.Len() != 0 {
		t.Errorf("expected 0, got %d", f.Len())
	}
}

func TestFlusher_Add_IncrementsLen(t *testing.T) {
	f := NewFlusher(DefaultFlushPolicy(), &bytes.Buffer{})
	r := DriftResult{Service: "svc-a"}
	if err := f.Add(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Len() != 1 {
		t.Errorf("expected 1, got %d", f.Len())
	}
}

func TestFlusher_Flush_WritesOutput(t *testing.T) {
	var buf bytes.Buffer
	f := NewFlusher(DefaultFlushPolicy(), &buf)
	f.Add(DriftResult{Service: "svc-a", Entries: []DriftEntry{{Kind: KindImage}}})
	if err := f.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "svc-a") {
		t.Errorf("expected output to contain service name, got: %s", out)
	}
	if !strings.Contains(out, "drifted") {
		t.Errorf("expected output to contain 'drifted', got: %s", out)
	}
}

func TestFlusher_Flush_ClearsBuffer(t *testing.T) {
	var buf bytes.Buffer
	f := NewFlusher(DefaultFlushPolicy(), &buf)
	f.Add(DriftResult{Service: "svc-a"})
	f.Flush()
	if f.Len() != 0 {
		t.Errorf("expected buffer cleared, got %d", f.Len())
	}
}

func TestFlusher_AutoFlush_OnMaxSize(t *testing.T) {
	var buf bytes.Buffer
	f := NewFlusher(FlushPolicy{MaxSize: 2, MaxAge: time.Hour}, &buf)
	f.Add(DriftResult{Service: "svc-a"})
	if f.Len() != 1 {
		t.Errorf("expected 1 buffered result")
	}
	f.Add(DriftResult{Service: "svc-b"})
	// auto-flush should have fired
	if f.Len() != 0 {
		t.Errorf("expected buffer flushed after MaxSize, got %d", f.Len())
	}
	if !strings.Contains(buf.String(), "2 result(s)") {
		t.Errorf("expected flush header with 2 results, got: %s", buf.String())
	}
}

func TestFlusher_Flush_Empty_NoError(t *testing.T) {
	var buf bytes.Buffer
	f := NewFlusher(DefaultFlushPolicy(), &buf)
	if err := f.Flush(); err != nil {
		t.Errorf("unexpected error on empty flush: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty flush")
	}
}
