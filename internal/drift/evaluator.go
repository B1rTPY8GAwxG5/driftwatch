package drift

import "fmt"

// EvaluationRule defines a named condition applied to a DriftResult.
type EvaluationRule struct {
	Name      string
	Condition func(DriftResult) bool
	Message   string
}

// EvaluationOutcome holds the result of applying all rules to a DriftResult.
type EvaluationOutcome struct {
	Service  string
	Passed   []string
	Failed   []string
	Blocked  bool
}

// HasFailures returns true if any rule failed.
func (o EvaluationOutcome) HasFailures() bool {
	return len(o.Failed) > 0
}

// Summary returns a human-readable summary of the outcome.
func (o EvaluationOutcome) Summary() string {
	if !o.HasFailures() {
		return fmt.Sprintf("service %s: all %d rule(s) passed", o.Service, len(o.Passed))
	}
	return fmt.Sprintf("service %s: %d rule(s) failed: %v", o.Service, len(o.Failed), o.Failed)
}

// Evaluator applies a set of named rules to DriftResults.
type Evaluator struct {
	rules []EvaluationRule
}

// NewEvaluator returns an Evaluator with no rules.
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// AddRule appends a rule to the evaluator.
func (e *Evaluator) AddRule(r EvaluationRule) {
	if r.Name == "" || r.Condition == nil {
		return
	}
	e.rules = append(e.rules, r)
}

// Evaluate applies all rules to the given result and returns an EvaluationOutcome.
func (e *Evaluator) Evaluate(result DriftResult) EvaluationOutcome {
	out := EvaluationOutcome{Service: result.Service}
	for _, rule := range e.rules {
		if rule.Condition(result) {
			out.Passed = append(out.Passed, rule.Name)
		} else {
			out.Failed = append(out.Failed, rule.Name)
		}
	}
	out.Blocked = out.HasFailures()
	return out
}

// EvaluateAll evaluates a slice of results and returns all outcomes.
func (e *Evaluator) EvaluateAll(results []DriftResult) []EvaluationOutcome {
	outs := make([]EvaluationOutcome, 0, len(results))
	for _, r := range results {
		outs = append(outs, e.Evaluate(r))
	}
	return outs
}
