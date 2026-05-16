package drift

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// QuotaConfig is the YAML-serialisable form of a QuotaPolicy.
type QuotaConfig struct {
	Limit  int    `yaml:"limit"`
	Period string `yaml:"period"`
}

// LoadQuotaConfig reads a QuotaConfig from a YAML file at the given path.
func LoadQuotaConfig(path string) (QuotaPolicy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return QuotaPolicy{}, fmt.Errorf("quota: read file: %w", err)
	}
	return LoadQuotaConfigFromBytes(data)
}

// LoadQuotaConfigFromBytes parses a QuotaPolicy from raw YAML bytes.
func LoadQuotaConfigFromBytes(data []byte) (QuotaPolicy, error) {
	var cfg QuotaConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return QuotaPolicy{}, fmt.Errorf("quota: parse yaml: %w", err)
	}
	if cfg.Limit <= 0 {
		return QuotaPolicy{}, fmt.Errorf("quota: limit must be greater than zero")
	}
	if cfg.Period == "" {
		return QuotaPolicy{}, fmt.Errorf("quota: period is required")
	}
	policy := QuotaPolicy{
		Limit:  cfg.Limit,
		Period: QuotaPeriod(cfg.Period),
	}
	if err := policy.Validate(); err != nil {
		return QuotaPolicy{}, fmt.Errorf("quota: invalid policy: %w", err)
	}
	return policy, nil
}

// BuildQuotaEnforcer loads a QuotaPolicy from a YAML file and returns a ready enforcer.
func BuildQuotaEnforcer(path string) (*QuotaEnforcer, error) {
	policy, err := LoadQuotaConfig(path)
	if err != nil {
		return nil, err
	}
	return NewQuotaEnforcer(policy)
}
