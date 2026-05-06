package drift

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func snapshotResult() DriftResult {
	return DriftResult{
		Service: "api-server",
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "nginx:1.24", Observed: "nginx:1.25"},
		},
	}
}

func TestNewSnapshotStore_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "snapshots")
	store, err := NewSnapshotStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("directory not created: %v", err)
	}
}

func TestSnapshotStore_SaveAndLoad(t *testing.T) {
	store, _ := NewSnapshotStore(t.TempDir())
	result := snapshotResult()

	if err := store.Save(result); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	snap, err := store.Load(result.Service)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if snap.Service != result.Service {
		t.Errorf("expected service %q, got %q", result.Service, snap.Service)
	}
	if len(snap.Result.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(snap.Result.Entries))
	}
	if snap.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestSnapshotStore_Load_NotFound(t *testing.T) {
	store, _ := NewSnapshotStore(t.TempDir())
	_, err := store.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing snapshot")
	}
}

func TestSnapshotStore_Exists_True(t *testing.T) {
	store, _ := NewSnapshotStore(t.TempDir())
	result := snapshotResult()
	_ = store.Save(result)

	if !store.Exists(result.Service) {
		t.Error("expected Exists to return true")
	}
}

func TestSnapshotStore_Exists_False(t *testing.T) {
	store, _ := NewSnapshotStore(t.TempDir())
	if store.Exists("missing-service") {
		t.Error("expected Exists to return false")
	}
}

func TestSnapshot_TimestampIsUTC(t *testing.T) {
	store, _ := NewSnapshotStore(t.TempDir())
	_ = store.Save(snapshotResult())
	snap, _ := store.Load("api-server")
	if snap.Timestamp.Location() != time.UTC {
		t.Errorf("expected UTC timestamp, got %v", snap.Timestamp.Location())
	}
}
