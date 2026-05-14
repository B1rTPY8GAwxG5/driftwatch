package drift

import (
	"os"
	"testing"
	"time"
)

func checkpointResult(service string, drifted bool) DriftResult {
	entries := []DriftEntry{}
	if drifted {
		entries = append(entries, DriftEntry{
			Kind:     KindImage,
			Field:    "image",
			Expected: "nginx:1.24",
			Actual:   "nginx:1.25",
		})
	}
	return DriftResult{Service: service, Entries: entries}
}

func TestNewCheckpointStore_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/checkpoints"
	store, err := NewCheckpointStore(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected dir to exist: %v", err)
	}
}

func TestCheckpointStore_SaveAndLoad(t *testing.T) {
	store, _ := NewCheckpointStore(t.TempDir())
	result := checkpointResult("api", true)

	if err := store.Save(result); err != nil {
		t.Fatalf("save: %v", err)
	}

	cp, err := store.Load("api")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cp.Service != "api" {
		t.Errorf("expected service api, got %s", cp.Service)
	}
	if !cp.Drifted {
		t.Error("expected drifted=true")
	}
	if cp.EntryCount != 1 {
		t.Errorf("expected entry_count=1, got %d", cp.EntryCount)
	}
	if cp.RunAt.IsZero() {
		t.Error("expected non-zero run_at")
	}
}

func TestCheckpointStore_Load_NotFound(t *testing.T) {
	store, _ := NewCheckpointStore(t.TempDir())
	_, err := store.Load("missing-service")
	if err == nil {
		t.Fatal("expected error for missing service")
	}
}

func TestCheckpointStore_Exists_True(t *testing.T) {
	store, _ := NewCheckpointStore(t.TempDir())
	result := checkpointResult("worker", false)
	_ = store.Save(result)
	if !store.Exists("worker") {
		t.Error("expected Exists=true")
	}
}

func TestCheckpointStore_Exists_False(t *testing.T) {
	store, _ := NewCheckpointStore(t.TempDir())
	if store.Exists("ghost") {
		t.Error("expected Exists=false")
	}
}

func TestCheckpointStore_Delete(t *testing.T) {
	store, _ := NewCheckpointStore(t.TempDir())
	result := checkpointResult("svc", false)
	_ = store.Save(result)
	if err := store.Delete("svc"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if store.Exists("svc") {
		t.Error("expected checkpoint to be deleted")
	}
}

func TestCheckpointStore_Delete_NotFound_NoError(t *testing.T) {
	store, _ := NewCheckpointStore(t.TempDir())
	if err := store.Delete("nonexistent"); err != nil {
		t.Errorf("expected no error for missing checkpoint, got: %v", err)
	}
}

func TestCheckpoint_Fields(t *testing.T) {
	cp := Checkpoint{
		Service:    "gateway",
		RunAt:      time.Now(),
		Drifted:    true,
		EntryCount: 3,
	}
	if cp.Service != "gateway" {
		t.Errorf("unexpected service: %s", cp.Service)
	}
	if cp.EntryCount != 3 {
		t.Errorf("unexpected entry count: %d", cp.EntryCount)
	}
}
