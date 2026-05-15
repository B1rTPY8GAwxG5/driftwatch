package drift

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

var changelogDriftedResult = DriftResult{
	Service: "api",
	Entries: []DriftEntry{
		{Kind: KindImage, Declared: "nginx:1.24", Observed: "nginx:1.25"},
		{Kind: KindReplicas, Declared: "3", Observed: "2"},
	},
}

var changelogCleanResult = DriftResult{
	Service: "worker",
	Entries: []DriftEntry{},
}

func TestNewChangelog_Empty(t *testing.T) {
	cl := NewChangelog()
	if cl == nil {
		t.Fatal("expected non-nil Changelog")
	}
	if cl.Len() != 0 {
		t.Errorf("expected 0 entries, got %d", cl.Len())
	}
}

func TestChangelog_Record_CleanResult_NoEntries(t *testing.T) {
	cl := NewChangelog()
	cl.Record(changelogCleanResult)
	if cl.Len() != 0 {
		t.Errorf("expected 0 entries for clean result, got %d", cl.Len())
	}
}

func TestChangelog_Record_DriftedResult_AddsEntries(t *testing.T) {
	cl := NewChangelog()
	cl.Record(changelogDriftedResult)
	if cl.Len() != 2 {
		t.Errorf("expected 2 entries, got %d", cl.Len())
	}
}

func TestChangelog_Entries_ReturnsCopy(t *testing.T) {
	cl := NewChangelog()
	cl.Record(changelogDriftedResult)
	a := cl.Entries()
	a[0].Service = "mutated"
	b := cl.Entries()
	if b[0].Service == "mutated" {
		t.Error("Entries should return a copy, not a reference")
	}
}

func TestChangelog_Entries_SortedByTime(t *testing.T) {
	cl := NewChangelog()
	cl.Record(changelogDriftedResult)
	entries := cl.Entries()
	for i := 1; i < len(entries); i++ {
		if entries[i].DetectedAt.Before(entries[i-1].DetectedAt) {
			t.Error("entries not sorted chronologically")
		}
	}
}

func TestChangelog_WriteTo_NoDrift(t *testing.T) {
	cl := NewChangelog()
	var buf bytes.Buffer
	if err := cl.WriteTo(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no drift") {
		t.Errorf("expected 'no drift' in output, got: %s", buf.String())
	}
}

func TestChangelog_WriteTo_WithDrift(t *testing.T) {
	cl := NewChangelog()
	cl.Record(changelogDriftedResult)
	var buf bytes.Buffer
	if err := cl.WriteTo(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "api") {
		t.Errorf("expected service name in output, got: %s", out)
	}
	if !strings.Contains(out, "nginx:1.25") {
		t.Errorf("expected observed image in output, got: %s", out)
	}
}

func TestChangelog_Entry_Fields(t *testing.T) {
	cl := NewChangelog()
	cl.Record(changelogDriftedResult)
	entries := cl.Entries()
	e := entries[0]
	if e.Service != "api" {
		t.Errorf("expected service 'api', got %q", e.Service)
	}
	if e.Kind != KindImage {
		t.Errorf("expected kind %v, got %v", KindImage, e.Kind)
	}
	if e.DetectedAt.IsZero() {
		t.Error("expected non-zero DetectedAt")
	}
	if e.DetectedAt.After(time.Now().UTC().Add(time.Second)) {
		t.Error("DetectedAt should not be in the future")
	}
}
