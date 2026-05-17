package drift

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMapperConfigFromBytes_Valid(t *testing.T) {
	data := []byte("mode: kind\n")
	cfg, err := LoadMapperConfigFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != "kind" {
		t.Errorf("expected mode 'kind', got %s", cfg.Mode)
	}
}

func TestLoadMapperConfigFromBytes_DefaultMode(t *testing.T) {
	data := []byte("{}\n")
	cfg, err := LoadMapperConfigFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != string(MapByService) {
		t.Errorf("expected default mode 'service', got %s", cfg.Mode)
	}
}

func TestLoadMapperConfigFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadMapperConfigFromBytes([]byte("::bad"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadMapperConfig_FileNotFound(t *testing.T) {
	_, err := LoadMapperConfig("/nonexistent/mapper.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadMapperConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mapper.yaml")
	_ = os.WriteFile(p, []byte("mode: service\n"), 0o644)
	cfg, err := LoadMapperConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != "service" {
		t.Errorf("expected 'service', got %s", cfg.Mode)
	}
}

func TestBuildMapper_FromConfig(t *testing.T) {
	cfg := &MapperConfig{Mode: "kind"}
	m := BuildMapper(cfg)
	if m.Mode() != MapByKind {
		t.Errorf("expected MapByKind, got %s", m.Mode())
	}
}
