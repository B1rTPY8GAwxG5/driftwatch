package drift

import (
	"os"
	"testing"
)

func TestNewRemediationStore_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/remediation"
	store, err := NewRemediationStore(path)
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

func TestRemediationStore_SaveAndLoad(t *testing.T) {
	store, _ := NewRemediationStore(t.TempDir())
	plan := BuildRemediationPlan(driftedForRemediation)

	if err := store.Save(plan); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load(plan.Service)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.Service != plan.Service {
		t.Errorf("service mismatch: got %q want %q", loaded.Service, plan.Service)
	}
	if len(loaded.Actions) != len(plan.Actions) {
		t.Errorf("action count mismatch: got %d want %d", len(loaded.Actions), len(plan.Actions))
	}
}

func TestRemediationStore_Load_NotFound(t *testing.T) {
	store, _ := NewRemediationStore(t.TempDir())
	_, err := store.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing plan")
	}
}

func TestRemediationStore_Exists_True(t *testing.T) {
	store, _ := NewRemediationStore(t.TempDir())
	plan := BuildRemediationPlan(driftedForRemediation)
	_ = store.Save(plan)
	if !store.Exists(plan.Service) {
		t.Error("expected Exists to return true")
	}
}

func TestRemediationStore_Exists_False(t *testing.T) {
	store, _ := NewRemediationStore(t.TempDir())
	if store.Exists("ghost") {
		t.Error("expected Exists to return false")
	}
}
