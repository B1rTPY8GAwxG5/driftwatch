package drift

import (
	"strings"
	"testing"
	"time"
)

func TestNewDriftProfiler_NotNil(t *testing.T) {
	p := NewDriftProfiler(0)
	if p == nil {
		t.Fatal("expected non-nil profiler")
	}
}

func TestNewDriftProfiler_DefaultMaxSize(t *testing.T) {
	p := NewDriftProfiler(0)
	if p.maxSize != 256 {
		t.Fatalf("expected maxSize 256, got %d", p.maxSize)
	}
}

func TestDriftProfiler_Record_AddsEntry(t *testing.T) {
	p := NewDriftProfiler(10)
	p.Record("svc-a", "compare", 5*time.Millisecond)
	entries := p.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Service != "svc-a" {
		t.Errorf("expected service svc-a, got %s", entries[0].Service)
	}
	if entries[0].Operation != "compare" {
		t.Errorf("expected operation compare, got %s", entries[0].Operation)
	}
}

func TestDriftProfiler_Record_EvictsOldestWhenFull(t *testing.T) {
	p := NewDriftProfiler(2)
	p.Record("svc-a", "op", 1*time.Millisecond)
	p.Record("svc-b", "op", 2*time.Millisecond)
	p.Record("svc-c", "op", 3*time.Millisecond)
	entries := p.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries after eviction, got %d", len(entries))
	}
	if entries[0].Service != "svc-b" {
		t.Errorf("expected svc-b after eviction, got %s", entries[0].Service)
	}
}

func TestDriftProfiler_TopN_SortedDescending(t *testing.T) {
	p := NewDriftProfiler(10)
	p.Record("fast", "op", 1*time.Millisecond)
	p.Record("slow", "op", 100*time.Millisecond)
	p.Record("mid", "op", 50*time.Millisecond)
	top := p.TopN(2)
	if len(top) != 2 {
		t.Fatalf("expected 2 top entries, got %d", len(top))
	}
	if top[0].Service != "slow" {
		t.Errorf("expected slow first, got %s", top[0].Service)
	}
	if top[1].Service != "mid" {
		t.Errorf("expected mid second, got %s", top[1].Service)
	}
}

func TestDriftProfiler_TopN_ClampsToLen(t *testing.T) {
	p := NewDriftProfiler(10)
	p.Record("svc", "op", 1*time.Millisecond)
	top := p.TopN(100)
	if len(top) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(top))
	}
}

func TestDriftProfiler_Summary_NoEntries(t *testing.T) {
	p := NewDriftProfiler(10)
	s := p.Summary()
	if !strings.Contains(s, "no entries") {
		t.Errorf("expected no entries message, got: %s", s)
	}
}

func TestDriftProfiler_Summary_WithEntries(t *testing.T) {
	p := NewDriftProfiler(10)
	p.Record("svc-a", "compare", 10*time.Millisecond)
	p.Record("svc-b", "compare", 20*time.Millisecond)
	s := p.Summary()
	if !strings.Contains(s, "entries=2") {
		t.Errorf("expected entries=2 in summary, got: %s", s)
	}
	if !strings.Contains(s, "svc-b") {
		t.Errorf("expected slowest service in summary, got: %s", s)
	}
}

func TestDriftProfiler_Entries_ReturnsCopy(t *testing.T) {
	p := NewDriftProfiler(10)
	p.Record("svc", "op", time.Millisecond)
	e1 := p.Entries()
	e1[0].Service = "mutated"
	e2 := p.Entries()
	if e2[0].Service == "mutated" {
		t.Error("Entries should return a copy, not a reference")
	}
}
