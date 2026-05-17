package drift

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// CompressionFormat identifies the compression algorithm to use.
type CompressionFormat string

const (
	CompressionNone CompressionFormat = "none"
	CompressionGzip CompressionFormat = "gzip"
)

// CompressorOptions configures the Compressor.
type CompressorOptions struct {
	Format CompressionFormat
	Level  int // gzip level; 0 means gzip.DefaultCompression
}

// DefaultCompressorOptions returns sensible defaults.
func DefaultCompressorOptions() CompressorOptions {
	return CompressorOptions{
		Format: CompressionGzip,
		Level:  gzip.DefaultCompression,
	}
}

// Compressor compresses and decompresses raw byte payloads.
type Compressor struct {
	opts CompressorOptions
}

// NewCompressor creates a Compressor with the given options.
// Zero-value options fall back to defaults.
func NewCompressor(opts CompressorOptions) *Compressor {
	if opts.Format == "" {
		opts = DefaultCompressorOptions()
	}
	if opts.Level == 0 {
		opts.Level = gzip.DefaultCompression
	}
	return &Compressor{opts: opts}
}

// Compress returns the compressed form of src.
func (c *Compressor) Compress(src []byte) ([]byte, error) {
	if c.opts.Format == CompressionNone {
		return src, nil
	}
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, c.opts.Level)
	if err != nil {
		return nil, fmt.Errorf("compressor: new writer: %w", err)
	}
	if _, err := w.Write(src); err != nil {
		return nil, fmt.Errorf("compressor: write: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("compressor: close: %w", err)
	}
	return buf.Bytes(), nil
}

// Decompress returns the decompressed form of src.
func (c *Compressor) Decompress(src []byte) ([]byte, error) {
	if c.opts.Format == CompressionNone {
		return src, nil
	}
	r, err := gzip.NewReader(bytes.NewReader(src))
	if err != nil {
		return nil, fmt.Errorf("compressor: new reader: %w", err)
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("compressor: read: %w", err)
	}
	return out, nil
}

// Format returns the active CompressionFormat.
func (c *Compressor) Format() CompressionFormat { return c.opts.Format }
