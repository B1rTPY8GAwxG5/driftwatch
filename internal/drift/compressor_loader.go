package drift

import (
	"compress/gzip"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// compressorConfig is the YAML-serialisable configuration for a Compressor.
type compressorConfig struct {
	Format string `yaml:"format"`
	Level  int    `yaml:"level"`
}

// LoadCompressorConfig reads a Compressor configuration from a YAML file.
func LoadCompressorConfig(path string) (*Compressor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load compressor config: %w", err)
	}
	return LoadCompressorConfigFromBytes(data)
}

// LoadCompressorConfigFromBytes parses a Compressor configuration from raw YAML.
func LoadCompressorConfigFromBytes(data []byte) (*Compressor, error) {
	var cfg compressorConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("load compressor config: unmarshal: %w", err)
	}

	opts := DefaultCompressorOptions()

	if cfg.Format != "" {
		switch CompressionFormat(cfg.Format) {
		case CompressionNone, CompressionGzip:
			opts.Format = CompressionFormat(cfg.Format)
		default:
			return nil, fmt.Errorf("load compressor config: unknown format %q", cfg.Format)
		}
	}

	if cfg.Level != 0 {
		if cfg.Level < gzip.HuffmanOnly || cfg.Level > gzip.BestCompression {
			return nil, fmt.Errorf("load compressor config: invalid level %d", cfg.Level)
		}
		opts.Level = cfg.Level
	}

	return NewCompressor(opts), nil
}
