package drift

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadClassifierConfigFromBytes_Valid(t *testing.T) {
	yaml := []byte(`
rules:
  - kind: image
    field: image
    category: container
  - kind: replicas
    category: scaling
`)
	c, err := LoadClassifierConfigFromBytes(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil classifier")
	}
}

func TestLoadClassifierConfigFromBytes_MissingKind(t *testing.T) {
	yaml := []byte(`
rules:
  - field: image
    category: container
`)
	_, err := LoadClassifierConfigFromBytes(yaml)
	if err == nil {
		t.Fatal("expected error for missing kind")
	}
}

func TestLoadClassifierConfigFromBytes_MissingCategory(t *testing.T) {
	yaml := []byte(`
rules:
  - kind: image
    field: image
`)
	_, err := LoadClassifierConfigFromBytes(yaml)
	if err == nil {
		t.Fatal("expected error for missing category")
	}
}

func TestLoadClassifierConfigFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadClassifierConfigFromBytes([]byte(":::bad yaml:::"))
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoadClassifierConfig_FileNotFound(t *testing.T) {
	_, err := LoadClassifierConfig("/nonexistent/classifier.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadClassifierConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "classifier.yaml")
	content := []byte(`
rules:
  - kind: image
    category: container
`)
	if err := os.WriteFile(p, content, 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadClassifierConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil classifier")
	}
}
