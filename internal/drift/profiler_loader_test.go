package drift

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProfilerConfigFromBytes_Valid(t *testing.T) {
	data := []byte("max_size: 64\n")
	cfg, err := LoadProfilerConfigFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxSize != 64 {
		t.Errorf("expected MaxSize 64, got %d", cfg.MaxSize)
	}
}

func TestLoadProfilerConfigFromBytes_ZeroMaxSize(t *testing.T) {
	data := []byte("max_size: 0\n")
	cfg, err := LoadProfilerConfigFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxSize != 0 {
		t.Errorf("expected MaxSize 0, got %d", cfg.MaxSize)
	}
}

func TestLoadProfilerConfigFromBytes_NegativeMaxSize(t *testing.T) {
	data := []byte("max_size: -1\n")
	_, err := LoadProfilerConfigFromBytes(data)
	if err == nil {
		t.Fatal("expected error for negative max_size")
	}
}

func TestLoadProfilerConfigFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadProfilerConfigFromBytes([]byte(":::invalid"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadProfilerConfig_FileNotFound(t *testing.T) {
	_, err := LoadProfilerConfig("/nonexistent/profiler.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadProfilerConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profiler.yaml")
	_ = os.WriteFile(path, []byte("max_size: 128\n"), 0o644)
	cfg, err := LoadProfilerConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxSize != 128 {
		t.Errorf("expected MaxSize 128, got %d", cfg.MaxSize)
	}
}

func TestBuildDriftProfiler_NotNil(t *testing.T) {
	cfg := ProfilerConfig{MaxSize: 32}
	p := BuildDriftProfiler(cfg)
	if p == nil {
		t.Fatal("expected non-nil profiler")
	}
	if p.maxSize != 32 {
		t.Errorf("expected maxSize 32, got %d", p.maxSize)
	}
}
