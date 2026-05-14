package drift

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWindowConfigFromBytes_Valid(t *testing.T) {
	data := []byte("size: 2m\nmax_items: 50\n")
	p, err := LoadWindowConfigFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.MaxItems != 50 {
		t.Errorf("expected 50, got %d", p.MaxItems)
	}
}

func TestLoadWindowConfigFromBytes_DefaultMaxItems(t *testing.T) {
	data := []byte("size: 10m\n")
	p, err := LoadWindowConfigFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.MaxItems != DefaultWindowPolicy().MaxItems {
		t.Errorf("expected default max_items, got %d", p.MaxItems)
	}
}

func TestLoadWindowConfigFromBytes_MissingSize(t *testing.T) {
	data := []byte("max_items: 10\n")
	_, err := LoadWindowConfigFromBytes(data)
	if err == nil {
		t.Error("expected error for missing size")
	}
}

func TestLoadWindowConfigFromBytes_InvalidDuration(t *testing.T) {
	data := []byte("size: notaduration\nmax_items: 10\n")
	_, err := LoadWindowConfigFromBytes(data)
	if err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestLoadWindowConfigFromBytes_InvalidYAML(t *testing.T) {
	data := []byte(": : :")
	_, err := LoadWindowConfigFromBytes(data)
	if err == nil {
		t.Error("expected error for invalid yaml")
	}
}

func TestLoadWindowConfig_FileNotFound(t *testing.T) {
	_, err := LoadWindowConfig("/nonexistent/window.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadWindowConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "window.yaml")
	_ = os.WriteFile(path, []byte("size: 3m\nmax_items: 20\n"), 0644)
	p, err := LoadWindowConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.MaxItems != 20 {
		t.Errorf("expected 20, got %d", p.MaxItems)
	}
}

func TestBuildSlidingWindow_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "window.yaml")
	_ = os.WriteFile(path, []byte("size: 1m\nmax_items: 10\n"), 0644)
	w, err := BuildSlidingWindow(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w == nil {
		t.Error("expected non-nil window")
	}
}
