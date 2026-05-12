package drift

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// BudgetConfig holds YAML-deserialised budget configuration.
type BudgetConfig struct {
	Limit  int    `yaml:"limit"`
	Period string `yaml:"period"`
}

// LoadBudgetConfig reads a BudgetConfig from a YAML file at path.
func LoadBudgetConfig(path string) (*BudgetConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("budget config: read file: %w", err)
	}
	return LoadBudgetConfigFromBytes(data)
}

// LoadBudgetConfigFromBytes parses a BudgetConfig from raw YAML bytes.
func LoadBudgetConfigFromBytes(data []byte) (*BudgetConfig, error) {
	var cfg BudgetConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("budget config: parse yaml: %w", err)
	}
	if cfg.Limit <= 0 {
		return nil, fmt.Errorf("budget config: limit must be positive, got %d", cfg.Limit)
	}
	if cfg.Period == "" {
		return nil, fmt.Errorf("budget config: period must not be empty")
	}
	return &cfg, nil
}

// BuildDriftBudget constructs a DriftBudget from a BudgetConfig.
func BuildDriftBudget(cfg *BudgetConfig) (*DriftBudget, error) {
	if cfg == nil {
		return nil, fmt.Errorf("budget config must not be nil")
	}
	return NewDriftBudget(cfg.Limit, BudgetPeriod(cfg.Period))
}
