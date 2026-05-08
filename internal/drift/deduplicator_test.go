package drift

import (
	"testing"
)

func TestNewDeduplicator_NotNil(t *testing.T) {
	d := NewDeduplicator()
	if d == nil {
		t.Fatal("expected non-nil Deduplicator")
	}
}

func TestDeduplicator_IsDuplicate_FirstCall_False(t *testing.T) {
	d := NewDeduplicator()
	k := DriftKey{Service: "svc", Kind: KindImage, Field: "image"}
	if d.IsDuplicate(k) {
		t.Error("first call should not be a duplicate")
	}
}

func TestDeduplicator_IsDuplicate_SecondCall_True(t *testing.T) {
	d := NewDeduplicator()
	k := DriftKey{Service: "svc", Kind: KindImage, Field: "image"}
	d.IsDuplicate(k)
	if !d.IsDuplicate(k) {
		t.Error("second call should be a duplicate")
	}
}

func TestDeduplicator_Reset_ClearsState(t *testing.T) {
	d := NewDeduplicator()
	k := DriftKey{Service: "svc", Kind: KindReplicas, Field: "replicas"}
	d.IsDuplicate(k)
	d.Reset()
	if d.IsDuplicate(k) {
		t.Error("after reset, key should not be a duplicate")
	}
}

func TestDeduplicator_Deduplicate_RemovesSeen(t *testing.T) {
	d := NewDeduplicator()
	result := DriftResult{
		Service: "api",
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "v1", Observed: "v2"},
			{Kind: KindReplicas, Field: "replicas", Declared: "2", Observed: "3"},
		},
	}
	// First pass: both entries are new.
	out := d.Deduplicate(result)
	if len(out.Entries) != 2 {
		t.Fatalf("expected 2 entries on first pass, got %d", len(out.Entries))
	}
	// Second pass: both entries are duplicates.
	out2 := d.Deduplicate(result)
	if len(out2.Entries) != 0 {
		t.Fatalf("expected 0 entries on second pass, got %d", len(out2.Entries))
	}
}

func TestDeduplicator_Deduplicate_PartialDuplicates(t *testing.T) {
	d := NewDeduplicator()
	first := DriftResult{
		Service: "svc",
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "v1", Observed: "v2"},
		},
	}
	d.Deduplicate(first)

	second := DriftResult{
		Service: "svc",
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "v1", Observed: "v2"},
			{Kind: KindReplicas, Field: "replicas", Declared: "1", Observed: "3"},
		},
	}
	out := d.Deduplicate(second)
	if len(out.Entries) != 1 {
		t.Fatalf("expected 1 new entry, got %d", len(out.Entries))
	}
	if out.Entries[0].Kind != KindReplicas {
		t.Errorf("expected KindReplicas, got %s", out.Entries[0].Kind)
	}
}

func TestDeduplicator_Deduplicate_PreservesServiceName(t *testing.T) {
	d := NewDeduplicator()
	result := DriftResult{
		Service: "my-service",
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "a", Observed: "b"},
		},
	}
	out := d.Deduplicate(result)
	if out.Service != "my-service" {
		t.Errorf("expected service 'my-service', got '%s'", out.Service)
	}
}
