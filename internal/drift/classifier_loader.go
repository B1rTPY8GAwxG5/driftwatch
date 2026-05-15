package drift

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// classifierRuleConfig is the YAML representation of a single classifier rule.
type classifierRuleConfig struct {
	Kind     string `yaml:"kind"`
	Field    string `yaml:"field"`
	Category string `yaml:"category"`
}

// classifierConfig is the top-level YAML structure for classifier configuration.
type classifierConfig struct {
	Rules []classifierRuleConfig `yaml:"rules"`
}

// LoadClassifierConfig reads a classifier configuration from a file path.
func LoadClassifierConfig(path string) (*Classifier, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("classifier: read file: %w", err)
	}
	return LoadClassifierConfigFromBytes(data)
}

// LoadClassifierConfigFromBytes parses YAML bytes into a Classifier.
func LoadClassifierConfigFromBytes(data []byte) (*Classifier, error) {
	var cfg classifierConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("classifier: parse yaml: %w", err)
	}
	rules := make([]ClassifierRule, 0, len(cfg.Rules))
	for _, r := range cfg.Rules {
		if r.Kind == "" {
			return nil, fmt.Errorf("classifier: rule missing kind")
		}
		if r.Category == "" {
			return nil, fmt.Errorf("classifier: rule missing category")
		}
		rules = append(rules, ClassifierRule{
			Kind:     DriftKind(r.Kind),
			Field:    r.Field,
			Category: r.Category,
		})
	}
	return NewClassifier(rules), nil
}
