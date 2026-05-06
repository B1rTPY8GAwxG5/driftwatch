package drift

import (
	"fmt"
	"strings"
)

// PolicyAction defines what to do when drift is detected.
type PolicyAction string

const (
	PolicyActionWarn PolicyAction = "warn"
	PolicyActionBlock PolicyAction = "block"
	PolicyActionIgnore PolicyAction = "ignore"
)

// PolicyRule maps a DriftKind to an action.
type PolicyRule struct {
	Kind   DriftKind    `yaml:"kind"`
	Action PolicyAction `yaml:"action"`
}

// Policy holds a set of rules evaluated against a DriftResult.
type Policy struct {
	Name  string       `yaml:"name"`
	Rules []PolicyRule `yaml:"rules"`
}

// PolicyViolation describes a rule that was triggered.
type PolicyViolation struct {
	Rule    PolicyRule
	Entries []DriftEntry
}

// Evaluate checks the result against all rules and returns any violations.
func (p *Policy) Evaluate(result DriftResult) []PolicyViolation {
	index := make(map[DriftKind][]DriftEntry)
	for _, e := range result.Entries {
		index[e.Kind] = append(index[e.Kind], e)
	}

	var violations []PolicyViolation
	for _, rule := range p.Rules {
		if rule.Action == PolicyActionIgnore {
			continue
		}
		if entries, ok := index[rule.Kind]; ok {
			violations = append(violations, PolicyViolation{Rule: rule, Entries: entries})
		}
	}
	return violations
}

// Blocked returns true if any violation has a block action.
func (p *Policy) Blocked(violations []PolicyViolation) bool {
	for _, v := range violations {
		if v.Rule.Action == PolicyActionBlock {
			return true
		}
	}
	return false
}

// Summary returns a human-readable summary of violations.
func (v PolicyViolation) Summary() string {
	fields := make([]string, 0, len(v.Entries))
	for _, e := range v.Entries {
		fields = append(fields, e.Field)
	}
	return fmt.Sprintf("[%s] %s: %s", v.Rule.Action, v.Rule.Kind, strings.Join(fields, ", "))
}
