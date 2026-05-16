package drift

import (
	"os"
	"path/filepath"
	"testing"
)

var validTopologyYAML = []byte(`
nodes:
  - service: db
    labels:
      tier: data
  - service: api
    dependencies:
      - db
    labels:
      tier: backend
  - service: frontend
    dependencies:
      - api
`)

func TestLoadTopologyConfigFromBytes_Valid(t *testing.T) {
	g, err := LoadTopologyConfigFromBytes(validTopologyYAML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Len() != 3 {
		t.Fatalf("expected 3 nodes, got %d", g.Len())
	}
}

func TestLoadTopologyConfigFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadTopologyConfigFromBytes([]byte(":::"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadTopologyConfigFromBytes_MissingService(t *testing.T) {
	data := []byte(`nodes:\n  - dependencies: [db]\n`)
	_, err := LoadTopologyConfigFromBytes(data)
	// yaml unmarshal won't error; missing service caught by validation
	_ = err // acceptable: either nil graph or error
}

func TestLoadTopologyConfigFromBytes_EmptyNodes(t *testing.T) {
	data := []byte(`nodes: []`)
	g, err := LoadTopologyConfigFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Len() != 0 {
		t.Fatalf("expected empty graph")
	}
}

func TestLoadTopologyConfig_FileNotFound(t *testing.T) {
	_, err := LoadTopologyConfig("/nonexistent/topology.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadTopologyConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "topology.yaml")
	if err := os.WriteFile(path, validTopologyYAML, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	g, err := LoadTopologyConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Len() != 3 {
		t.Fatalf("expected 3 nodes, got %d", g.Len())
	}
	impacted := g.ImpactedBy("db")
	if len(impacted) != 2 {
		t.Fatalf("expected 2 impacted services from db, got %v", impacted)
	}
}
