package drift

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ObserverConfig holds YAML-decoded configuration for an Observer.
type ObserverConfig struct {
	BufferEvents bool `yaml:"buffer_events"`
	MaxEvents    int  `yaml:"max_events"`
}

// LoadObserverConfig reads an ObserverConfig from the given file path.
func LoadObserverConfig(path string) (ObserverConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ObserverConfig{}, fmt.Errorf("observer: read file: %w", err)
	}
	return LoadObserverConfigFromBytes(data)
}

// LoadObserverConfigFromBytes parses an ObserverConfig from raw YAML bytes.
func LoadObserverConfigFromBytes(data []byte) (ObserverConfig, error) {
	var cfg ObserverConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return ObserverConfig{}, fmt.Errorf("observer: parse yaml: %w", err)
	}
	if cfg.MaxEvents < 0 {
		return ObserverConfig{}, fmt.Errorf("observer: max_events must be non-negative")
	}
	return cfg, nil
}

// BuildObserver constructs an Observer from the provided config.
func BuildObserver(cfg ObserverConfig) *Observer {
	o := NewObserver()
	_ = cfg // reserved for future buffer/cap enforcement
	return o
}
