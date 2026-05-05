package drift

import (
	"fmt"
	"io"
	"os"
)

// SinkType identifies the destination of a notification sink.
type SinkType string

const (
	SinkStdout SinkType = "stdout"
	SinkStderr SinkType = "stderr"
	SinkFile   SinkType = "file"
)

// SinkConfig describes how to construct a notification sink writer.
type SinkConfig struct {
	Type SinkType `yaml:"type"`
	Path string   `yaml:"path,omitempty"`
}

// OpenSink returns an io.WriteCloser for the given SinkConfig.
// Callers are responsible for closing file-based sinks.
func OpenSink(cfg SinkConfig) (io.WriteCloser, error) {
	switch cfg.Type {
	case SinkStdout:
		return nopCloser{os.Stdout}, nil
	case SinkStderr:
		return nopCloser{os.Stderr}, nil
	case SinkFile:
		if cfg.Path == "" {
			return nil, fmt.Errorf("sink type 'file' requires a path")
		}
		f, err := os.OpenFile(cfg.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, fmt.Errorf("opening sink file %q: %w", cfg.Path, err)
		}
		return f, nil
	default:
		return nil, fmt.Errorf("unknown sink type %q", cfg.Type)
	}
}

// nopCloser wraps a writer that needs no closing.
type nopCloser struct{ io.Writer }

func (nopCloser) Close() error { return nil }
