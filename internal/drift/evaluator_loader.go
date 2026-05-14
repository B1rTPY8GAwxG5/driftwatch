package drift

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// EvaluatorRuleConfig is the YAML representation of a single evaluation rule.
type EvaluatorRuleConfig struct {
	Name string `yaml:"name"`
	Kind string `yaml:"kind"` // "no_drift", "max_entries"
	Max  int    `yaml:"max"`  // used by max_entries
}

// EvaluatorConfig is the top-level YAML structure for evaluator rules.
type EvaluatorConfig struct {
	Rules []EvaluatorRuleConfig `yaml:"rules"`
}

// LoadEvaluatorConfig reads an EvaluatorConfig from a YAML file.
func LoadEvaluatorConfig(path string) (*EvaluatorConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("evaluator config: %w", err)
	}
	return LoadEvaluatorConfigFromBytes(data)
}

// LoadEvaluatorConfigFromBytes parses an EvaluatorConfig from raw YAML bytes.
func LoadEvaluatorConfigFromBytes(data []byte) (*EvaluatorConfig, error) {
	var cfg EvaluatorConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("evaluator config: invalid yaml: %w", err)
	}
	for _, r := range cfg.Rules {
		if r.Name == "" {
			return nil, fmt.Errorf("evaluator config: rule missing name")
		}
	}
	return &cfg, nil
}

// BuildEvaluator constructs an Evaluator from a config, registering built-in rule kinds.
func BuildEvaluator(cfg *EvaluatorConfig) (*Evaluator, error) {
	e := NewEvaluator()
	for _, rc := range cfg.Rules {
		rule, err := builtinRule(rc)
		if err != nil {
			return nil, err
		}
		e.AddRule(rule)
	}
	return e, nil
}

func builtinRule(rc EvaluatorRuleConfig) (EvaluationRule, error) {
	switch rc.Kind {
	case "no_drift":
		return EvaluationRule{
			Name:      rc.Name,
			Condition: func(r DriftResult) bool { return !r.HasDrift() },
			Message:   "service must have no drift",
		}, nil
	case "max_entries":
		max := rc.Max
		return EvaluationRule{
			Name:      rc.Name,
			Condition: func(r DriftResult) bool { return len(r.Entries) <= max },
			Message:   fmt.Sprintf("drift entries must not exceed %d", max),
		}, nil
	default:
		return EvaluationRule{}, fmt.Errorf("evaluator: unknown rule kind %q", rc.Kind)
	}
}
