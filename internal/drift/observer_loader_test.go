package drift

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadObserverConfigFromBytes_Valid(t *testing.T) {
	data := []byte(`buffer_events: true\nmax_events: 100\n`)
	cfg, err := LoadObserverConfigFromBytes([]byte("buffer_events: true\nmax_events: 100\n"))
	_ = data
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.BufferEvents {
		t.Error("expected buffer_events true")
	}
	if cfg.MaxEvents != 100 {
		t.Errorf("expected max_events 100, got %d", cfg.MaxEvents)
	}
}

func TestLoadObserverConfigFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadObserverConfigFromBytes([]byte(":::invalid"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadObserverConfigFromBytes_NegativeMaxEvents(t *testing.T) {
	_, err := LoadObserverConfigFromBytes([]byte("max_events: -1\n"))
	if err == nil {
		t.Fatal("expected error for negative max_events")
	}
}

func TestLoadObserverConfig_FileNotFound(t *testing.T) {
	_, err := LoadObserverConfig("/nonexistent/observer.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadObserverConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "observer.yaml")
	content := []byte("buffer_events: false\nmax_events: 50\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	cfg, err := LoadObserverConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxEvents != 50 {
		t.Errorf("expected 50, got %d", cfg.MaxEvents)
	}
}

func TestBuildObserver_NotNil(t *testing.T) {
	cfg := ObserverConfig{BufferEvents: true, MaxEvents: 10}
	o := BuildObserver(cfg)
	if o == nil {
		t.Fatal("expected non-nil observer")
	}
}
