package drift

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCompressorConfigFromBytes_Valid(t *testing.T) {
	yaml := []byte(`format: gzip\nlevel: 6\n`)
	c, err := LoadCompressorConfigFromBytes(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil Compressor")
	}
	if c.Format() != CompressionGzip {
		t.Errorf("expected gzip, got %s", c.Format())
	}
}

func TestLoadCompressorConfigFromBytes_NoneFormat(t *testing.T) {
	yaml := []byte(`format: none\n`)
	c, err := LoadCompressorConfigFromBytes(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Format() != CompressionNone {
		t.Errorf("expected none, got %s", c.Format())
	}
}

func TestLoadCompressorConfigFromBytes_UnknownFormat(t *testing.T) {
	yaml := []byte(`format: zstd\n`)
	_, err := LoadCompressorConfigFromBytes(yaml)
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestLoadCompressorConfigFromBytes_InvalidLevel(t *testing.T) {
	yaml := []byte(`format: gzip\nlevel: 99\n`)
	_, err := LoadCompressorConfigFromBytes(yaml)
	if err == nil {
		t.Fatal("expected error for invalid level")
	}
}

func TestLoadCompressorConfigFromBytes_InvalidYAML(t *testing.T) {
	_, err := LoadCompressorConfigFromBytes([]byte(":::bad yaml:::"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadCompressorConfigFromBytes_EmptyUsesDefaults(t *testing.T) {
	c, err := LoadCompressorConfigFromBytes([]byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Format() != CompressionGzip {
		t.Errorf("expected default gzip, got %s", c.Format())
	}
	if c.opts.Level != gzip.DefaultCompression {
		t.Errorf("expected default level, got %d", c.opts.Level)
	}
}

func TestLoadCompressorConfig_FileNotFound(t *testing.T) {
	_, err := LoadCompressorConfig("/nonexistent/compressor.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadCompressorConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "compressor.yaml")
	_ = os.WriteFile(p, []byte(`format: gzip\nlevel: 1\n`), 0o644)
	c, err := LoadCompressorConfig(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.opts.Level != 1 {
		t.Errorf("expected level 1, got %d", c.opts.Level)
	}
}
