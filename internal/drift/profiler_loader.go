package drift

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ProfilerConfig holds configuration for constructing a DriftProfiler.
type ProfilerConfig struct {
	MaxSize int `yaml:"max_size"`
}

// LoadProfilerConfig reads a ProfilerConfig from a YAML file at path.
func LoadProfilerConfig(path string) (ProfilerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ProfilerConfig{}, fmt.Errorf("profiler config: read file: %w", err)
	}
	return LoadProfilerConfigFromBytes(data)
}

// LoadProfilerConfigFromBytes parses a ProfilerConfig from raw YAML bytes.
func LoadProfilerConfigFromBytes(data []byte) (ProfilerConfig, error) {
	var cfg ProfilerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return ProfilerConfig{}, fmt.Errorf("profiler config: parse yaml: %w", err)
	}
	if cfg.MaxSize < 0 {
		return ProfilerConfig{}, fmt.Errorf("profiler config: max_size must be non-negative")
	}
	return cfg, nil
}

// BuildDriftProfiler constructs a DriftProfiler from a ProfilerConfig.
func BuildDriftProfiler(cfg ProfilerConfig) *DriftProfiler {
	return NewDriftProfiler(cfg.MaxSize)
}
