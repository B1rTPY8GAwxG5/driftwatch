package drift

import (
	"os"
	"path/filepath"
	"testing"
)

const validSpecYAML = `
name: my-service
image: nginx:1.25
replicas: 3
env:
  LOG_LEVEL: info
  PORT: "8080"
ports:
  - 80
  - 443
`

func TestLoadSpecFromBytes_Valid(t *testing.T) {
	spec, err := LoadSpecFromBytes([]byte(validSpecYAML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Name != "my-service" {
		t.Errorf("expected name 'my-service', got %q", spec.Name)
	}
	if spec.Image != "nginx:1.25" {
		t.Errorf("expected image 'nginx:1.25', got %q", spec.Image)
	}
	if spec.Replicas != 3 {
		t.Errorf("expected replicas 3, got %d", spec.Replicas)
	}
	if spec.Env["LOG_LEVEL"] != "info" {
		t.Errorf("expected LOG_LEVEL=info, got %q", spec.Env["LOG_LEVEL"])
	}
	if len(spec.Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(spec.Ports))
	}
}

func TestLoadSpecFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadSpecFromBytes([]byte(":::invalid yaml:::"))
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadSpec_FileNotFound(t *testing.T) {
	_, err := LoadSpec("/nonexistent/path/spec.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadSpec_ValidFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "spec.yaml")
	if err := os.WriteFile(path, []byte(validSpecYAML), 0644); err != nil {
		t.Fatalf("failed to write temp spec: %v", err)
	}

	spec, err := LoadSpec(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Name != "my-service" {
		t.Errorf("expected name 'my-service', got %q", spec.Name)
	}
}

func TestLoadSpec_MissingName(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "spec.yaml")
	data := []byte("image: nginx:1.25\nreplicas: 1\n")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write temp spec: %v", err)
	}

	_, err := LoadSpec(path)
	if err == nil {
		t.Fatal("expected error for missing name field, got nil")
	}
}
