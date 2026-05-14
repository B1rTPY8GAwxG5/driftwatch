package drift

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type windowConfig struct {
	Size     string `yaml:"size"`
	MaxItems int    `yaml:"max_items"`
}

// LoadWindowConfig reads a WindowPolicy from a YAML file at path.
func LoadWindowConfig(path string) (WindowPolicy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return WindowPolicy{}, fmt.Errorf("window config: %w", err)
	}
	return LoadWindowConfigFromBytes(data)
}

// LoadWindowConfigFromBytes parses a WindowPolicy from raw YAML bytes.
func LoadWindowConfigFromBytes(data []byte) (WindowPolicy, error) {
	var cfg windowConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return WindowPolicy{}, fmt.Errorf("window config: invalid yaml: %w", err)
	}
	if cfg.Size == "" {
		return WindowPolicy{}, fmt.Errorf("window config: size is required")
	}
	d, err := time.ParseDuration(cfg.Size)
	if err != nil {
		return WindowPolicy{}, fmt.Errorf("window config: invalid size %q: %w", cfg.Size, err)
	}
	maxItems := cfg.MaxItems
	if maxItems == 0 {
		maxItems = DefaultWindowPolicy().MaxItems
	}
	p := WindowPolicy{Size: d, MaxItems: maxItems}
	if err := p.Validate(); err != nil {
		return WindowPolicy{}, fmt.Errorf("window config: %w", err)
	}
	return p, nil
}

// BuildSlidingWindow constructs a SlidingWindow from a YAML config file.
func BuildSlidingWindow(path string) (*SlidingWindow, error) {
	p, err := LoadWindowConfig(path)
	if err != nil {
		return nil, err
	}
	return NewSlidingWindow(p)
}
