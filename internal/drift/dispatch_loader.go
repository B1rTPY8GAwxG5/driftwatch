package drift

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// DispatchConfig holds YAML-based configuration for a Dispatcher.
type DispatchConfig struct {
	Mode     string   `yaml:"mode"`     // "serial" or "parallel"
	Sinks    []string `yaml:"sinks"`    // sink targets (stdout, stderr, file path)
	OnDrift  bool     `yaml:"on_drift"` // only dispatch when drift detected
}

// LoadDispatchConfig reads a DispatchConfig from a file path.
func LoadDispatchConfig(path string) (*DispatchConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("dispatch loader: read file: %w", err)
	}
	return LoadDispatchConfigFromBytes(data)
}

// LoadDispatchConfigFromBytes parses a DispatchConfig from raw YAML bytes.
func LoadDispatchConfigFromBytes(data []byte) (*DispatchConfig, error) {
	var cfg DispatchConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("dispatch loader: unmarshal: %w", err)
	}
	if cfg.Mode == "" {
		cfg.Mode = "serial"
	}
	if cfg.Mode != "serial" && cfg.Mode != "parallel" {
		return nil, fmt.Errorf("dispatch loader: unknown mode %q", cfg.Mode)
	}
	return &cfg, nil
}

// BuildDispatcher constructs a Dispatcher from a DispatchConfig.
func BuildDispatcher(cfg *DispatchConfig) (*Dispatcher, error) {
	if cfg == nil {
		return nil, fmt.Errorf("dispatch loader: nil config")
	}
	mode := DispatchSerial
	if cfg.Mode == "parallel" {
		mode = DispatchParallel
	}
	d := NewDispatcher(mode, nil)
	return d, nil
}
