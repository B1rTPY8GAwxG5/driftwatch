package drift

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProjectionConfigFromBytes_Valid(t *testing.T) {
	yaml := []byte(`fields: [service, kind, drifted]`)
	cfg, err := LoadProjectionConfigFromBytes(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(cfg.Fields))
	}
}

func TestLoadProjectionConfigFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadProjectionConfigFromBytes([]byte(":::bad yaml"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadProjectionConfigFromBytes_EmptyFields(t *testing.T) {
	yaml := []byte(`fields: []`)
	cfg, err := LoadProjectionConfigFromBytes(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Fields) != 0 {
		t.Errorf("expected 0 fields")
	}
}

func TestLoadProjectionConfig_FileNotFound(t *testing.T) {
	_, err := LoadProjectionConfig("/no/such/file.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadProjectionConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "proj.yaml")
	content := []byte(`fields: [service, actual]`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	cfg, err := LoadProjectionConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(cfg.Fields))
	}
}

func TestBuildProjection_KnownFields(t *testing.T) {
	cfg := &ProjectionConfig{Fields: []string{"service", "kind", "drifted"}}
	p := BuildProjection(cfg)
	if len(p.Headers()) != 3 {
		t.Errorf("expected 3 headers, got %d", len(p.Headers()))
	}
}

func TestBuildProjection_UnknownFieldsIgnored(t *testing.T) {
	cfg := &ProjectionConfig{Fields: []string{"service", "nonexistent", "kind"}}
	p := BuildProjection(cfg)
	if len(p.Headers()) != 2 {
		t.Errorf("expected 2 headers (unknown ignored), got %d", len(p.Headers()))
	}
}

func TestBuildProjection_EmptyConfig_AllFields(t *testing.T) {
	cfg := &ProjectionConfig{Fields: []string{}}
	p := BuildProjection(cfg)
	// empty fields list → NewProjection() with no args → all 7 defaults
	if len(p.Headers()) != 7 {
		t.Errorf("expected 7 default headers, got %d", len(p.Headers()))
	}
}
