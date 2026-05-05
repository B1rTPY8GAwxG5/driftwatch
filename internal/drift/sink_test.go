package drift_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/driftwatch/internal/drift"
)

func TestOpenSink_Stdout(t *testing.T) {
	cfg := drift.SinkConfig{Type: drift.SinkStdout}
	wc, err := drift.OpenSink(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer wc.Close()
	if wc == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestOpenSink_Stderr(t *testing.T) {
	cfg := drift.SinkConfig{Type: drift.SinkStderr}
	wc, err := drift.OpenSink(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer wc.Close()
}

func TestOpenSink_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "drift.log")
	cfg := drift.SinkConfig{Type: drift.SinkFile, Path: path}
	wc, err := drift.OpenSink(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer wc.Close()
	_, err = wc.Write([]byte("test\n"))
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "test\n" {
		t.Errorf("unexpected file content: %q", string(data))
	}
}

func TestOpenSink_File_MissingPath(t *testing.T) {
	cfg := drift.SinkConfig{Type: drift.SinkFile}
	_, err := drift.OpenSink(cfg)
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestOpenSink_Unknown(t *testing.T) {
	cfg := drift.SinkConfig{Type: "webhook"}
	_, err := drift.OpenSink(cfg)
	if err == nil {
		t.Fatal("expected error for unknown sink type")
	}
}
