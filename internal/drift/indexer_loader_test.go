package drift

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadIndexerConfigFromBytes_Valid(t *testing.T) {
	data := []byte("mode: kind\n")
	cfg, err := LoadIndexerConfigFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != "kind" {
		t.Errorf("expected kind, got %s", cfg.Mode)
	}
}

func TestLoadIndexerConfigFromBytes_DefaultMode(t *testing.T) {
	data := []byte("{}")
	cfg, err := LoadIndexerConfigFromBytes(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != string(IndexByService) {
		t.Errorf("expected service default, got %s", cfg.Mode)
	}
}

func TestLoadIndexerConfigFromBytes_UnknownMode(t *testing.T) {
	data := []byte("mode: bogus\n")
	_, err := LoadIndexerConfigFromBytes(data)
	if err == nil {
		t.Fatal("expected error for unknown mode")
	}
}

func TestLoadIndexerConfigFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadIndexerConfigFromBytes([]byte("::::"))
	if err == nil {
		t.Fatal("expected error for invalid yaml")
	}
}

func TestLoadIndexerConfig_FileNotFound(t *testing.T) {
	_, err := LoadIndexerConfig("/no/such/file.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadIndexerConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "indexer.yaml")
	if err := os.WriteFile(p, []byte("mode: both\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadIndexerConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != "both" {
		t.Errorf("expected both, got %s", cfg.Mode)
	}
}

func TestBuildIndexer_NilConfig_UsesDefault(t *testing.T) {
	idx := BuildIndexer(nil)
	if idx == nil {
		t.Fatal("expected non-nil indexer")
	}
	if idx.Mode() != string(IndexByService) {
		t.Errorf("expected service mode, got %s", idx.Mode())
	}
}

func TestBuildIndexer_ValidConfig(t *testing.T) {
	cfg := &IndexerConfig{Mode: "kind"}
	idx := BuildIndexer(cfg)
	if idx.Mode() != "kind" {
		t.Errorf("expected kind mode, got %s", idx.Mode())
	}
}
