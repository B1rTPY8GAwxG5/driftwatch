package drift

import (
	"os"
	"testing"
)

var baselineSpec = ServiceSpec{
	Name:     "api",
	Image:    "api:v1.0.0",
	Replicas: 3,
	Env:      map[string]string{"PORT": "8080"},
}

func TestNewBaselineStore_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	store, err := NewBaselineStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestBaselineStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewBaselineStore(dir)

	if err := store.Save("api", baselineSpec); err != nil {
		t.Fatalf("Save: %v", err)
	}

	b, err := store.Load("api")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Service != "api" {
		t.Errorf("expected service api, got %s", b.Service)
	}
	if b.Spec.Image != baselineSpec.Image {
		t.Errorf("expected image %s, got %s", baselineSpec.Image, b.Spec.Image)
	}
	if b.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestBaselineStore_Load_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewBaselineStore(dir)

	_, err := store.Load("missing")
	if err == nil {
		t.Fatal("expected error for missing baseline")
	}
}

func TestBaselineStore_Exists_True(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewBaselineStore(dir)
	_ = store.Save("api", baselineSpec)

	if !store.Exists("api") {
		t.Error("expected Exists to return true")
	}
}

func TestBaselineStore_Exists_False(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewBaselineStore(dir)

	if store.Exists("ghost") {
		t.Error("expected Exists to return false")
	}
}

func TestBaselineStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewBaselineStore(dir)
	_ = store.Save("api", baselineSpec)

	if err := store.Delete("api"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if store.Exists("api") {
		t.Error("expected baseline to be deleted")
	}
}

func TestBaselineStore_Delete_NotFound_NoError(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewBaselineStore(dir)

	if err := store.Delete("nonexistent"); err != nil {
		t.Errorf("expected no error deleting nonexistent baseline, got: %v", err)
	}
}

func TestNewBaselineStore_InvalidDir(t *testing.T) {
	// Use a file as the directory path to force an error.
	f, _ := os.CreateTemp("", "baseline-test")
	defer os.Remove(f.Name())
	f.Close()

	_, err := NewBaselineStore(f.Name() + "/subdir")
	if err == nil {
		t.Fatal("expected error for invalid dir")
	}
}
