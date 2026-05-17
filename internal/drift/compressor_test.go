package drift

import (
	"compress/gzip"
	"strings"
	"testing"
)

func TestDefaultCompressorOptions_Values(t *testing.T) {
	opts := DefaultCompressorOptions()
	if opts.Format != CompressionGzip {
		t.Errorf("expected gzip, got %s", opts.Format)
	}
	if opts.Level != gzip.DefaultCompression {
		t.Errorf("expected default compression level, got %d", opts.Level)
	}
}

func TestNewCompressor_NotNil(t *testing.T) {
	c := NewCompressor(DefaultCompressorOptions())
	if c == nil {
		t.Fatal("expected non-nil Compressor")
	}
}

func TestNewCompressor_ZeroOptions_UsesDefaults(t *testing.T) {
	c := NewCompressor(CompressorOptions{})
	if c.Format() != CompressionGzip {
		t.Errorf("expected gzip fallback, got %s", c.Format())
	}
}

func TestCompressor_Format_ReturnsFormat(t *testing.T) {
	c := NewCompressor(CompressorOptions{Format: CompressionNone})
	if c.Format() != CompressionNone {
		t.Errorf("expected none, got %s", c.Format())
	}
}

func TestCompressor_Compress_Decompress_RoundTrip(t *testing.T) {
	c := NewCompressor(DefaultCompressorOptions())
	orig := []byte("hello driftwatch configuration drift payload")
	compressed, err := c.Compress(orig)
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	if len(compressed) == 0 {
		t.Fatal("expected non-empty compressed output")
	}
	got, err := c.Decompress(compressed)
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	if string(got) != string(orig) {
		t.Errorf("round-trip mismatch: got %q, want %q", got, orig)
	}
}

func TestCompressor_None_PassThrough(t *testing.T) {
	c := NewCompressor(CompressorOptions{Format: CompressionNone})
	orig := []byte("plain text")
	out, err := c.Compress(orig)
	if err != nil {
		t.Fatalf("compress none: %v", err)
	}
	if string(out) != string(orig) {
		t.Errorf("expected pass-through, got %q", out)
	}
	out2, err := c.Decompress(orig)
	if err != nil {
		t.Fatalf("decompress none: %v", err)
	}
	if string(out2) != string(orig) {
		t.Errorf("expected pass-through decompress, got %q", out2)
	}
}

func TestCompressor_Compress_ReducesSize(t *testing.T) {
	c := NewCompressor(DefaultCompressorOptions())
	// repetitive content compresses well
	orig := []byte(strings.Repeat("drift-entry:", 200))
	compressed, err := c.Compress(orig)
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	if len(compressed) >= len(orig) {
		t.Errorf("expected compressed size < original; got %d vs %d", len(compressed), len(orig))
	}
}

func TestCompressor_Decompress_InvalidData_ReturnsError(t *testing.T) {
	c := NewCompressor(DefaultCompressorOptions())
	_, err := c.Decompress([]byte("not gzip data"))
	if err == nil {
		t.Fatal("expected error decompressing invalid data")
	}
}
