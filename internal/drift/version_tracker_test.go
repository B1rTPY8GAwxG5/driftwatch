package drift

import (
	"testing"
	"time"
)

func trackerResult(service, got string, kind DriftKind) DriftResult {
	return DriftResult{
		Service: service,
		Entries: []DriftEntry{
			{Kind: kind, Want: "old", Got: got},
		},
	}
}

func TestNewVersionTracker_NotNil(t *testing.T) {
	vt := NewVersionTracker(0)
	if vt == nil {
		t.Fatal("expected non-nil VersionTracker")
	}
}

func TestNewVersionTracker_DefaultMaxAge(t *testing.T) {
	vt := NewVersionTracker(0)
	if vt.maxAge != 24*time.Hour {
		t.Errorf("expected 24h default, got %v", vt.maxAge)
	}
}

func TestVersionTracker_Record_AddsEntry(t *testing.T) {
	vt := NewVersionTracker(time.Hour)
	r := trackerResult("svc-a", "v2", KindImage)
	vt.Record(r)

	h := vt.History("svc-a", string(KindImage))
	if len(h) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(h))
	}
	if h[0].Value != "v2" {
		t.Errorf("expected value v2, got %s", h[0].Value)
	}
}

func TestVersionTracker_History_ReturnsCopy(t *testing.T) {
	vt := NewVersionTracker(time.Hour)
	vt.Record(trackerResult("svc-a", "v1", KindImage))

	h := vt.History("svc-a", string(KindImage))
	h[0].Value = "mutated"

	h2 := vt.History("svc-a", string(KindImage))
	if h2[0].Value == "mutated" {
		t.Error("History should return a copy, not a reference")
	}
}

func TestVersionTracker_HasChanged_NoHistory_False(t *testing.T) {
	vt := NewVersionTracker(time.Hour)
	if vt.HasChanged("svc-a", string(KindImage), "v1") {
		t.Error("expected false when no history exists")
	}
}

func TestVersionTracker_HasChanged_SameValue_False(t *testing.T) {
	vt := NewVersionTracker(time.Hour)
	vt.Record(trackerResult("svc-a", "v1", KindImage))
	if vt.HasChanged("svc-a", string(KindImage), "v1") {
		t.Error("expected false when value unchanged")
	}
}

func TestVersionTracker_HasChanged_DifferentValue_True(t *testing.T) {
	vt := NewVersionTracker(time.Hour)
	vt.Record(trackerResult("svc-a", "v1", KindImage))
	if !vt.HasChanged("svc-a", string(KindImage), "v2") {
		t.Error("expected true when value changed")
	}
}

func TestVersionTracker_Evict_RemovesOldEntries(t *testing.T) {
	vt := NewVersionTracker(10 * time.Millisecond)
	vt.Record(trackerResult("svc-a", "v1", KindImage))

	time.Sleep(30 * time.Millisecond)
	// Trigger eviction by recording a new result.
	vt.Record(trackerResult("svc-b", "v1", KindImage))

	h := vt.History("svc-a", string(KindImage))
	if len(h) != 0 {
		t.Errorf("expected evicted entries, got %d", len(h))
	}
}

func TestVersionTracker_MultipleServices_Independent(t *testing.T) {
	vt := NewVersionTracker(time.Hour)
	vt.Record(trackerResult("svc-a", "v1", KindImage))
	vt.Record(trackerResult("svc-b", "v2", KindImage))

	ha := vt.History("svc-a", string(KindImage))
	hb := vt.History("svc-b", string(KindImage))

	if len(ha) != 1 || ha[0].Value != "v1" {
		t.Errorf("unexpected history for svc-a: %+v", ha)
	}
	if len(hb) != 1 || hb[0].Value != "v2" {
		t.Errorf("unexpected history for svc-b: %+v", hb)
	}
}
