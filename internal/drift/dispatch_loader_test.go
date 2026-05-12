package drift

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDispatchConfigFromBytes_SerialDefault(t *testing.T) {
	data := []byte(`sinks: [stdout]`)
	cfg, err := LoadDispatchConfigFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != "serial" {
		t.Errorf("expected serial, got %q", cfg.Mode)
	}
}

func TestLoadDispatchConfigFromBytes_Parallel(t *testing.T) {
	data := []byte(`mode: parallel\nsinks: [stdout]`)
	cfg, err := LoadDispatchConfigFromBytes([]byte("mode: parallel\nsinks:\n  - stdout\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != "parallel" {
		t.Errorf("expected parallel, got %q", cfg.Mode)
	}
	_ = data
}

func TestLoadDispatchConfigFromBytes_InvalidMode(t *testing.T) {
	data := []byte("mode: batch\n")
	_, err := LoadDispatchConfigFromBytes(data)
	if err == nil {
		t.Error("expected error for unknown mode")
	}
}

func TestLoadDispatchConfigFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadDispatchConfigFromBytes([]byte("::::"))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadDispatchConfig_FileNotFound(t *testing.T) {
	_, err := LoadDispatchConfig("/nonexistent/dispatch.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadDispatchConfig_ValidFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "dispatch.yaml")
	content := "mode: serial\nsinks:\n  - stdout\non_drift: true\n"
	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadDispatchConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.OnDrift {
		t.Error("expected on_drift to be true")
	}
}

func TestBuildDispatcher_NilConfig(t *testing.T) {
	_, err := BuildDispatcher(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestBuildDispatcher_SerialMode(t *testing.T) {
	cfg := &DispatchConfig{Mode: "serial"}
	d, err := BuildDispatcher(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Error("expected non-nil dispatcher")
	}
	if d.mode != DispatchSerial {
		t.Errorf("expected DispatchSerial")
	}
}

func TestBuildDispatcher_ParallelMode(t *testing.T) {
	cfg := &DispatchConfig{Mode: "parallel"}
	d, err := BuildDispatcher(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.mode != DispatchParallel {
		t.Errorf("expected DispatchParallel")
	}
}
