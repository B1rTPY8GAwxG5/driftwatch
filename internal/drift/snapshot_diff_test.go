package drift

import "testing"

func imageEntry() DriftEntry {
	return DriftEntry{Kind: KindImage, Field: "image", Declared: "nginx:1.24", Observed: "nginx:1.25"}
}

func replicasEntry() DriftEntry {
	return DriftEntry{Kind: KindReplicas, Field: "replicas", Declared: "3", Observed: "2"}
}

func TestCompareSnapshots_NoDiff(t *testing.T) {
	prev := DriftResult{Service: "svc", Entries: []DriftEntry{imageEntry()}}
	curr := DriftResult{Service: "svc", Entries: []DriftEntry{imageEntry()}}

	diff := CompareSnapshots(prev, curr)
	if diff.HasChanges() {
		t.Errorf("expected no changes, got resolved=%d new=%d", len(diff.Resolved), len(diff.New))
	}
}

func TestCompareSnapshots_NewEntry(t *testing.T) {
	prev := DriftResult{Service: "svc", Entries: []DriftEntry{imageEntry()}}
	curr := DriftResult{Service: "svc", Entries: []DriftEntry{imageEntry(), replicasEntry()}}

	diff := CompareSnapshots(prev, curr)
	if len(diff.New) != 1 {
		t.Errorf("expected 1 new entry, got %d", len(diff.New))
	}
	if len(diff.Resolved) != 0 {
		t.Errorf("expected 0 resolved, got %d", len(diff.Resolved))
	}
}

func TestCompareSnapshots_ResolvedEntry(t *testing.T) {
	prev := DriftResult{Service: "svc", Entries: []DriftEntry{imageEntry(), replicasEntry()}}
	curr := DriftResult{Service: "svc", Entries: []DriftEntry{imageEntry()}}

	diff := CompareSnapshots(prev, curr)
	if len(diff.Resolved) != 1 {
		t.Errorf("expected 1 resolved entry, got %d", len(diff.Resolved))
	}
	if len(diff.New) != 0 {
		t.Errorf("expected 0 new, got %d", len(diff.New))
	}
}

func TestSnapshotDiff_HasChanges_False(t *testing.T) {
	sd := SnapshotDiff{Service: "svc"}
	if sd.HasChanges() {
		t.Error("expected HasChanges to be false")
	}
}

func TestSnapshotDiff_Summary_NoChanges(t *testing.T) {
	sd := SnapshotDiff{Service: "svc"}
	s := sd.Summary()
	if s == "" {
		t.Error("expected non-empty summary")
	}
}

func TestSnapshotDiff_Summary_WithChanges(t *testing.T) {
	sd := SnapshotDiff{
		Service:  "svc",
		New:      []DriftEntry{imageEntry()},
		Resolved: []DriftEntry{replicasEntry()},
	}
	s := sd.Summary()
	if s == "" {
		t.Error("expected non-empty summary")
	}
}
