package drift

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// IndexerConfig holds YAML-decoded configuration for a DriftIndex.
type IndexerConfig struct {
	Mode string `yaml:"mode"`
}

// LoadIndexerConfig reads an IndexerConfig from a YAML file at path.
func LoadIndexerConfig(path string) (*IndexerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("indexer: read file: %w", err)
	}
	return LoadIndexerConfigFromBytes(data)
}

// LoadIndexerConfigFromBytes parses an IndexerConfig from raw YAML bytes.
func LoadIndexerConfigFromBytes(data []byte) (*IndexerConfig, error) {
	var cfg IndexerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("indexer: parse yaml: %w", err)
	}
	if cfg.Mode == "" {
		cfg.Mode = string(IndexByService)
	}
	switch IndexMode(cfg.Mode) {
	case IndexByService, IndexByKind, IndexByBoth:
	default:
		return nil, fmt.Errorf("indexer: unknown mode %q", cfg.Mode)
	}
	return &cfg, nil
}

// BuildIndexer constructs a DriftIndex from an IndexerConfig.
func BuildIndexer(cfg *IndexerConfig) *DriftIndex {
	if cfg == nil {
		return NewIndexer(IndexByService)
	}
	return NewIndexer(IndexMode(cfg.Mode))
}
